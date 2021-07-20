package expr

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/components/gtime"
	"github.com/grafana/grafana/pkg/expr/mathexp"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// Command is an interface for all expression commands.
type Command interface {
	NeedsVars() []string
	Execute(c context.Context, vars mathexp.Vars) (mathexp.Results, error)
}

// MathCommand is a command for a math expression such as "1 + $GA / 2"
type MathCommand struct {
	RawExpression string
	Expression    *mathexp.Expr
	refID         string
}

// NewMathCommand creates a new MathCommand. It will return an error
// if there is an error parsing expr.
func NewMathCommand(refID, expr string) (*MathCommand, error) {
	parsedExpr, err := mathexp.New(expr)
	if err != nil {
		return nil, err
	}
	return &MathCommand{
		RawExpression: expr,
		Expression:    parsedExpr,
		refID:         refID,
	}, nil
}

// UnmarshalMathCommand creates a MathCommand from Grafana's frontend query.
func UnmarshalMathCommand(rn *rawNode) (*MathCommand, error) {
	rawExpr, ok := rn.Query["expression"]
	if !ok {
		return nil, fmt.Errorf("math command for refId %v is missing an expression", rn.RefID)
	}
	exprString, ok := rawExpr.(string)
	if !ok {
		return nil, fmt.Errorf("expected math command for refId %v expression to be a string, got %T", rn.RefID, rawExpr)
	}

	gm, err := NewMathCommand(rn.RefID, exprString)
	if err != nil {
		return nil, fmt.Errorf("invalid math command type in '%v': %v", rn.RefID, err)
	}
	return gm, nil
}

// NeedsVars returns the variable names (refIds) that are dependencies
// to execute the command and allows the command to fulfill the Command interface.
func (gm *MathCommand) NeedsVars() []string {
	return gm.Expression.VarNames
}

// Execute runs the command and returns the results or an error if the command
// failed to execute.
func (gm *MathCommand) Execute(ctx context.Context, vars mathexp.Vars) (mathexp.Results, error) {
	return gm.Expression.Execute(gm.refID, vars)
}

// ReduceCommand is an expression command for reduction of a timeseries such as a min, mean, or max.
type ReduceCommand struct {
	Reducer     string
	VarToReduce string
	refID       string
}

// NewReduceCommand creates a new ReduceCMD.
func NewReduceCommand(refID, reducer, varToReduce string) *ReduceCommand {
	// TODO: validate reducer here, before execution
	return &ReduceCommand{
		Reducer:     reducer,
		VarToReduce: varToReduce,
		refID:       refID,
	}
}

// UnmarshalReduceCommand creates a MathCMD from Grafana's frontend query.
func UnmarshalReduceCommand(rn *rawNode) (*ReduceCommand, error) {
	rawVar, ok := rn.Query["expression"]
	if !ok {
		return nil, fmt.Errorf("no variable specified to reduce for refId %v", rn.RefID)
	}
	varToReduce, ok := rawVar.(string)
	if !ok {
		return nil, fmt.Errorf("expected reduce variable to be a string, got %T for refId %v", rawVar, rn.RefID)
	}
	varToReduce = strings.TrimPrefix(varToReduce, "$")

	rawReducer, ok := rn.Query["reducer"]
	if !ok {
		return nil, fmt.Errorf("no reducer specified for refId %v", rn.RefID)
	}
	redFunc, ok := rawReducer.(string)
	if !ok {
		return nil, fmt.Errorf("expected reducer to be a string, got %T for refId %v", rawReducer, rn.RefID)
	}

	return NewReduceCommand(rn.RefID, redFunc, varToReduce), nil
}

// NeedsVars returns the variable names (refIds) that are dependencies
// to execute the command and allows the command to fulfill the Command interface.
func (gr *ReduceCommand) NeedsVars() []string {
	return []string{gr.VarToReduce}
}

// Execute runs the command and returns the results or an error if the command
// failed to execute.
func (gr *ReduceCommand) Execute(ctx context.Context, vars mathexp.Vars) (mathexp.Results, error) {
	newRes := mathexp.Results{}
	for _, val := range vars[gr.VarToReduce].Values {
		series, ok := val.(mathexp.Series)
		if !ok {
			return newRes, fmt.Errorf("can only reduce type series, got type %v", val.Type())
		}
		num, err := series.Reduce(gr.refID, gr.Reducer)
		if err != nil {
			return newRes, err
		}
		newRes.Values = append(newRes.Values, num)
	}
	return newRes, nil
}

// ResampleCommand is an expression command for resampling of a timeseries.
type ResampleCommand struct {
	Window        time.Duration
	VarToResample string
	Downsampler   string
	Upsampler     string
	TimeRange     TimeRange
	refID         string
}

// NewResampleCommand creates a new ResampleCMD.
func NewResampleCommand(refID, rawWindow, varToResample string, downsampler string, upsampler string, tr TimeRange) (*ResampleCommand, error) {
	// TODO: validate reducer here, before execution
	window, err := gtime.ParseDuration(rawWindow)
	if err != nil {
		return nil, fmt.Errorf(`failed to parse resample "window" duration field %q: %w`, window, err)
	}
	return &ResampleCommand{
		Window:        window,
		VarToResample: varToResample,
		Downsampler:   downsampler,
		Upsampler:     upsampler,
		TimeRange:     tr,
		refID:         refID,
	}, nil
}

// UnmarshalResampleCommand creates a ResampleCMD from Grafana's frontend query.
func UnmarshalResampleCommand(rn *rawNode) (*ResampleCommand, error) {
	rawVar, ok := rn.Query["expression"]
	if !ok {
		return nil, fmt.Errorf("no variable to resample for refId %v", rn.RefID)
	}
	varToReduce, ok := rawVar.(string)
	if !ok {
		return nil, fmt.Errorf("expected resample input variable to be type string, but got type %T for refId %v", rawVar, rn.RefID)
	}
	varToReduce = strings.TrimPrefix(varToReduce, "$")
	varToResample := varToReduce

	rawWindow, ok := rn.Query["window"]
	if !ok {
		return nil, fmt.Errorf("no time duration specified for the window in resample command for refId %v", rn.RefID)
	}
	window, ok := rawWindow.(string)
	if !ok {
		return nil, fmt.Errorf("expected resample window to be a string, got %T for refId %v", rawWindow, rn.RefID)
	}

	rawDownsampler, ok := rn.Query["downsampler"]
	if !ok {
		return nil, fmt.Errorf("no downsampler function specified in resample command for refId %v", rn.RefID)
	}
	downsampler, ok := rawDownsampler.(string)
	if !ok {
		return nil, fmt.Errorf("expected resample downsampler to be a string, got type %T for refId %v", downsampler, rn.RefID)
	}

	rawUpsampler, ok := rn.Query["upsampler"]
	if !ok {
		return nil, fmt.Errorf("no downsampler specified in resample command for refId %v", rn.RefID)
	}
	upsampler, ok := rawUpsampler.(string)
	if !ok {
		return nil, fmt.Errorf("expected resample downsampler to be a string, got type %T for refId %v", upsampler, rn.RefID)
	}

	return NewResampleCommand(rn.RefID, window, varToResample, downsampler, upsampler, rn.TimeRange)
}

// NeedsVars returns the variable names (refIds) that are dependencies
// to execute the command and allows the command to fulfill the Command interface.
func (gr *ResampleCommand) NeedsVars() []string {
	return []string{gr.VarToResample}
}

// Execute runs the command and returns the results or an error if the command
// failed to execute.
func (gr *ResampleCommand) Execute(ctx context.Context, vars mathexp.Vars) (mathexp.Results, error) {
	newRes := mathexp.Results{}
	for _, val := range vars[gr.VarToResample].Values {
		series, ok := val.(mathexp.Series)
		if !ok {
			return newRes, fmt.Errorf("can only resample type series, got type %v", val.Type())
		}
		num, err := series.Resample(gr.refID, gr.Window, gr.Downsampler, gr.Upsampler, gr.TimeRange.From, gr.TimeRange.To)
		if err != nil {
			return newRes, err
		}
		newRes.Values = append(newRes.Values, num)
	}
	return newRes, nil
}

// Select MetricCommand is an expresion for selecting a specific metric name from a response.
// All labeled items matching the metric name are returned, so this for queries that not only return multiple
// items that are differentiated not only by label name but also by metric name.
// If IsRegex is true, then Metric name must be a valid regular expression.
type FilterItems struct {
	refID string

	InputVar string

	MetricName string

	LabelMatchers string
	matchers      []*labels.Matcher

	IsRegex bool
	re      *regexp.Regexp
}

// UnmarshalResampleCommand creates a ResampleCMD from Grafana's frontend query.
func UnmarshalFilterItemsCommand(rn *rawNode) (*FilterItems, error) {
	fi := &FilterItems{
		refID: rn.RefID,
	}

	rawInput, ok := rn.Query["expression"]
	if !ok {
		return nil, fmt.Errorf("no variable to select metric for refId %v", rn.RefID)
	}
	inputVar, ok := rawInput.(string)
	if !ok {
		return nil, fmt.Errorf("expected select metric input variable to be type string, but got type %T for refId %v", inputVar, rn.RefID)
	}
	fi.InputVar = strings.TrimPrefix(inputVar, "$")

	rawMetricName, ok := rn.Query["metricName"]
	if ok {
		fi.MetricName, ok = rawMetricName.(string)
		if !ok {
			return nil, fmt.Errorf("expected metric name in select metric to be a string, but got %T for refId %v", rawMetricName, rn.RefID)
		}
	}

	rawLabelMatchers, ok := rn.Query["labelMatchers"]
	if ok {
		fi.LabelMatchers, ok = rawLabelMatchers.(string)
		if !ok {
			return nil, fmt.Errorf("expected labelMatchers in select metric to be a string, but got %T for refId %v", rawLabelMatchers, rn.RefID)
		}
	}

	if fi.LabelMatchers == "" && fi.MetricName == "" {
		return nil, fmt.Errorf("no metric name or labels matcher specificed for select metric in refId %v", rn.RefID)
	}

	rawIsRegex, ok := rn.Query["isRegex"]
	if ok {
		fi.IsRegex, ok = rawIsRegex.(bool)
		if !ok {
			return nil, fmt.Errorf("expected isRegex in select metric to be a bool, but got %T for refId %v", rawMetricName, rn.RefID)
		}
	}

	if err := fi.buildFilters(); err != nil {
		return nil, err
	}

	return fi, nil
}

func (fi *FilterItems) buildFilters() error {
	if fi.LabelMatchers != "" {
		var err error
		fi.matchers, err = parser.ParseMetricSelector(fmt.Sprintf("{%v}", fi.LabelMatchers))
		if err != nil {
			return fmt.Errorf("invalid label matching string in select metric for refId %v: %w", fi.refID, err)
		}
	}
	if fi.IsRegex {
		var err error
		fi.re, err = regexp.Compile(fi.MetricName)
		if err != nil {
			return fmt.Errorf("invalid regular expression in select metric for refId %v: %w", fi.refID, err)
		}
	}
	return nil
}

// Execute runs the command and returns the results or an error if the command
// failed to execute.
func (fi *FilterItems) Execute(ctx context.Context, vars mathexp.Vars) (mathexp.Results, error) {
	newRes := mathexp.Results{}
	inputData := vars[fi.InputVar]

	appendValue := func(v mathexp.Value) {
		// new series/numbers need to be created for selection, but in this
		// filtering case we duplicates pointers
		switch val := v.(type) {
		case mathexp.Number:
			num := mathexp.NewNumber(fi.refID, val.GetLabels())
			num.SetValue(val.GetFloat64Value())
			newRes.Values = append(newRes.Values, num)
		case mathexp.Series:
			s := mathexp.NewSeries(fi.refID, val.GetLabels(), val.Len())
			// Note: sharing a reference to a Field's values between fields
			// is no possible since that slice is not exposed in data.Field,
			// so must do a new slices.
			for i := 0; i < val.Len(); i++ {
				t, f := val.GetPoint(i)
				s.SetPoint(i, t, f)
			}
			newRes.Values = append(newRes.Values, s)
		}
	}

	ifMatchAppend := func(metricName string, dl data.Labels, v mathexp.Value) {
		var matched bool
		pl := labels.FromMap(dl)
		if metricName != "" {
			if fi.IsRegex {
				matched = fi.re.MatchString(metricName)
			} else {
				matched = metricName == fi.MetricName
			}
		}
		if len(fi.matchers) > 0 {
			if labels.Selector(fi.matchers).Matches(pl) {
				if fi.MetricName != "" && !matched {
					matched = false
				} else {
					matched = true
				}
			} else {
				if fi.MetricName != "" && matched {
					matched = false
				}
			}
		}

		if matched {
			appendValue(v)
		}
	}

	for _, val := range inputData.Values {
		switch v := val.(type) {
		case mathexp.Series:
			ifMatchAppend(v.GetName(), v.GetLabels(), v)
		case mathexp.Number:
			ifMatchAppend(v.GetName(), v.GetLabels(), v)
		default:
			return newRes, fmt.Errorf("metric select input must be type Series or Number, but got type %s in refId %v", v.Type(), fi.refID)
		}
	}
	return newRes, nil
}

// NeedsVars returns the variable names (refIds) that are dependencies
// to execute the command and allows the command to fulfill the Command interface.
func (fi *FilterItems) NeedsVars() []string {
	return []string{fi.InputVar}
}

// CommandType is the type of the expression command.
type CommandType int

const (
	// TypeUnknown is the CMDType for an unrecognized expression type.
	TypeUnknown CommandType = iota
	// TypeMath is the CMDType for a math expression.
	TypeMath
	// TypeReduce is the CMDType for a reduction expression.
	TypeReduce
	// TypeResample is the CMDType for a resampling expression.
	TypeResample
	// TypeClassicConditions is the CMDType for the classic condition operation.
	TypeClassicConditions
	// TypeFilterItems is the CMDType for the select metric operation.
	TypeFilterItems
)

func (gt CommandType) String() string {
	switch gt {
	case TypeMath:
		return "math"
	case TypeReduce:
		return "reduce"
	case TypeResample:
		return "resample"
	case TypeClassicConditions:
		return "classic_conditions"
	case TypeFilterItems:
		return "filter_items"
	default:
		return "unknown"
	}
}

// ParseCommandType returns a CommandType from its string representation.
func ParseCommandType(s string) (CommandType, error) {
	switch s {
	case "math":
		return TypeMath, nil
	case "reduce":
		return TypeReduce, nil
	case "resample":
		return TypeResample, nil
	case "classic_conditions":
		return TypeClassicConditions, nil
	case "filter_items":
		return TypeFilterItems, nil
	default:
		return TypeUnknown, fmt.Errorf("'%v' is not a recognized expression type", s)
	}
}
