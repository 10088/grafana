package tsdb

import (
	"context"
	"fmt"

	"github.com/grafana/grafana/pkg/infra/httpclient"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/tsdb/azuremonitor"
	"github.com/grafana/grafana/pkg/tsdb/cloudmonitoring"
	"github.com/grafana/grafana/pkg/tsdb/cloudwatch"
	"github.com/grafana/grafana/pkg/tsdb/elasticsearch"
	"github.com/grafana/grafana/pkg/tsdb/graphite"
	"github.com/grafana/grafana/pkg/tsdb/influxdb"
	"github.com/grafana/grafana/pkg/tsdb/loki"
	"github.com/grafana/grafana/pkg/tsdb/mssql"
	"github.com/grafana/grafana/pkg/tsdb/mysql"
	"github.com/grafana/grafana/pkg/tsdb/opentsdb"
	"github.com/grafana/grafana/pkg/tsdb/postgres"
	"github.com/grafana/grafana/pkg/tsdb/prometheus"
	"github.com/grafana/grafana/pkg/tsdb/tempo"
)

// NewService returns a new Service.
func NewService(cfg *setting.Cfg, cloudWatchService *cloudwatch.CloudWatchService,
	cloudMonitoringService *cloudmonitoring.Service, azureMonitorService *azuremonitor.Service,
	pluginManager plugins.Manager, postgresService *postgres.PostgresService,
	httpClientProvider httpclient.Provider) *Service {
	return &Service{
		Cfg:                    cfg,
		CloudWatchService:      cloudWatchService,
		CloudMonitoringService: cloudMonitoringService,
		AzureMonitorService:    azureMonitorService,
		PluginManager:          pluginManager,
		registry: map[string]func(*models.DataSource) (plugins.DataPlugin, error){
			"graphite":                         graphite.New(httpClientProvider),
			"opentsdb":                         opentsdb.New(httpClientProvider),
			"prometheus":                       prometheus.New(httpClientProvider),
			"influxdb":                         influxdb.New(httpClientProvider),
			"mssql":                            mssql.NewExecutor,
			"postgres":                         postgresService.NewExecutor,
			"mysql":                            mysql.New(httpClientProvider),
			"elasticsearch":                    elasticsearch.New(httpClientProvider),
			"stackdriver":                      cloudMonitoringService.NewExecutor,
			"grafana-azure-monitor-datasource": azureMonitorService.NewExecutor,
			"loki":                             loki.New(httpClientProvider),
			"tempo":                            tempo.New(httpClientProvider),
		},
	}
}

// Service handles data requests to data sources.
type Service struct {
	Cfg                    *setting.Cfg
	CloudWatchService      *cloudwatch.CloudWatchService
	CloudMonitoringService *cloudmonitoring.Service
	AzureMonitorService    *azuremonitor.Service
	PluginManager          plugins.Manager

	//nolint: staticcheck // plugins.DataPlugin deprecated
	registry map[string]func(*models.DataSource) (plugins.DataPlugin, error)
}

// Init initialises the service.
func (s *Service) Init() error {
	return nil
}

//nolint: staticcheck // plugins.DataPlugin deprecated
func (s *Service) HandleRequest(ctx context.Context, ds *models.DataSource, query plugins.DataQuery) (
	plugins.DataResponse, error) {
	plugin := s.PluginManager.GetDataPlugin(ds.Type)
	if plugin == nil {
		factory, exists := s.registry[ds.Type]
		if !exists {
			return plugins.DataResponse{}, fmt.Errorf(
				"could not find plugin corresponding to data source type: %q", ds.Type)
		}

		var err error
		plugin, err = factory(ds)
		if err != nil {
			return plugins.DataResponse{}, fmt.Errorf("could not instantiate endpoint for data plugin %q: %w",
				ds.Type, err)
		}
	}

	return plugin.DataQuery(ctx, ds, query)
}

// RegisterQueryHandler registers a query handler factory.
// This is only exposed for tests!
//nolint: staticcheck // plugins.DataPlugin deprecated
func (s *Service) RegisterQueryHandler(name string, factory func(*models.DataSource) (plugins.DataPlugin, error)) {
	s.registry[name] = factory
}
