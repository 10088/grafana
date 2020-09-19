import { css } from 'emotion';
import { GrafanaTheme } from '@grafana/data';
import { styleMixins, stylesFactory } from '../../themes';
import { getScrollbarWidth } from '../../utils';
import { ComponentSize } from '../../types/size';

export interface TableStyles {
  cellHeight: number;
  cellHeightInner: number;
  cellPadding: number;
  rowHeight: number;
  table: string;
  thead: string;
  headerCell: string;
  headerCellLabel: string;
  headerFilter: string;
  tableCell: string;
  tableCellWrapper: string;
  tableCellLink: string;
  row: string;
  theme: GrafanaTheme;
  resizeHandle: string;
  overflow: string;
}

export const getTableStyles = stylesFactory(
  (theme: GrafanaTheme, cellSize: ComponentSize): TableStyles => {
    const { palette, colors } = theme;
    const headerBg = theme.colors.bg2;
    const borderColor = theme.colors.border1;
    const resizerColor = theme.isLight ? palette.blue95 : palette.blue77;

    let padding = 6;
    let lineHeight = theme.typography.lineHeight.md;
    let bodyFontSize = 14;
    let cellHeight = 34;

    if (cellSize === 'lg') {
      bodyFontSize = 16;
      cellHeight = 45;
    }

    const rowHoverBg = styleMixins.hoverColor(theme.colors.bg1, theme);
    const scollbarWidth = getScrollbarWidth();

    return {
      theme,
      cellHeight,
      cellPadding: padding,
      cellHeightInner: bodyFontSize * lineHeight,
      rowHeight: cellHeight + 2,
      table: css`
        height: 100%;
        width: 100%;
        overflow: auto;
        display: flex;
      `,
      thead: css`
        label: thead;
        height: ${cellHeight}px;
        overflow-y: auto;
        overflow-x: hidden;
        background: ${headerBg};
        position: relative;
      `,
      headerCell: css`
        padding: ${padding}px;
        overflow: hidden;
        white-space: nowrap;
        color: ${colors.textBlue};
        border-right: 1px solid ${theme.colors.panelBg};
        display: flex;

        &:last-child {
          border-right: none;
        }
      `,
      headerCellLabel: css`
        cursor: pointer;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        display: flex;
        margin-right: ${theme.spacing.xs};
      `,
      headerFilter: css`
        label: headerFilter;
        cursor: pointer;
      `,
      row: css`
        label: row;
        border-bottom: 1px solid ${borderColor};

        &:hover {
          background-color: ${rowHoverBg};
        }
      `,
      tableCellWrapper: css`
        border-right: 1px solid ${borderColor};
        display: inline-flex;
        align-items: center;
        height: 100%;

        &:last-child {
          border-right: none;

          > div {
            padding-right: ${scollbarWidth + padding}px;
          }
        }
      `,
      tableCellLink: css`
        text-decoration: underline;
      `,
      tableCell: css`
        padding: ${padding}px;
        text-overflow: ellipsis;
        white-space: nowrap;
        overflow: hidden;
        flex: 1;
        width: 100%;
        height: 100%;

        &:hover {
          overflow: visible;
          width: auto;
          box-shadow: 0 0 2px ${theme.colors.formFocusOutline};
          background: ${rowHoverBg};
          z-index: 1;
        }
      `,
      overflow: css`
        overflow: hidden;
        text-overflow: ellipsis;
      `,
      resizeHandle: css`
        label: resizeHandle;
        cursor: col-resize !important;
        display: inline-block;
        background: ${resizerColor};
        opacity: 0;
        transition: opacity 0.2s ease-in-out;
        width: 8px;
        height: 100%;
        position: absolute;
        right: -4px;
        border-radius: 3px;
        top: 0;
        z-index: ${theme.zIndex.dropdown};
        touch-action: none;

        &:hover {
          opacity: 1;
        }
      `,
    };
  }
);
