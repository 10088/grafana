import React from 'react';
import uPlot from 'uplot';
import { DataFrame, FieldColor, TimeRange, TimeZone } from '@grafana/data';

export type NullValuesMode = 'null' | 'connected' | 'asZero';

export enum AxisSide {
  Top,
  Right,
  Bottom,
  Left,
}

interface AxisConfig {
  label: string;
  side: AxisSide;
  grid: boolean;
  width: number;
}

interface LineConfig {
  show: boolean;
  width: number;
  color: FieldColor;
}
interface PointConfig {
  show: boolean;
  radius: number;
}
interface BarsConfig {
  show: boolean;
}
interface FillConfig {
  alpha: number;
}

export interface GraphCustomFieldConfig {
  axis: AxisConfig;
  line: LineConfig;
  points: PointConfig;
  bars: BarsConfig;
  fill: FillConfig;
  nullValues: NullValuesMode;
}

export type PlotSeriesConfig = Pick<uPlot.Options, 'series' | 'scales' | 'axes'>;
export type PlotPlugin = {
  id: string;
  /** can mutate provided opts as necessary */
  opts?: (self: uPlot, opts: uPlot.Options) => void;
  hooks: uPlot.PluginHooks;
};

export interface PlotPluginProps {
  id: string;
}

export interface PlotProps {
  data: DataFrame;
  timeRange: TimeRange;
  timeZone: TimeZone;
  width: number;
  height: number;
  config: PlotSeriesConfig;
  children?: React.ReactElement[];
  /** Callback performed when uPlot data is updated */
  onDataUpdate?: (data: AlignedData) => {};
  /** Callback performed when uPlot is (re)initialized */
  onPlotInit?: () => {};
}

export type AlignedData = [
  // x values
  number[],
  // y values
  ...Array<number | null>
];

// uPlot.Series.isGap ?
export type isGap = (self: uPlot, seriesIdx: number, dataIdx: number) => boolean;

export interface AlignedDataWithGapTest {
  data: AlignedData;
  isGap: isGap;
}

export interface AlignedFrameWithGapTest {
  frame: DataFrame;
  isGap: isGap;
}
