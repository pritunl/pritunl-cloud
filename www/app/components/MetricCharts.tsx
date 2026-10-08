/// <reference path="../References.d.ts"/>
import * as React from 'react';
import PageSelect from './PageSelect';
import MetricChart from './MetricChart';
import * as InstanceActions from '../actions/InstanceActions';
import * as NodeActions from '../actions/NodeActions';
import * as InstanceTypes from '../types/InstanceTypes';
import SearchInput from './SearchInput';

interface Props {
	instance?: string;
	node?: string;
	components?: InstanceTypes.Component[];
	disabled: boolean;
}

interface State {
	sync: number;
	period: number;
	interval: number;
	loading: {[key: string]: boolean};
	cancelable: {[key: string]: boolean};
	componentsFilter: string;
}

const css = {
	header: {
		fontSize: '20px',
		marginTop: '-10px',
		paddingBottom: '2px',
		marginBottom: '10px',
		borderBottomStyle: 'solid',
	} as React.CSSProperties,
	heading: {
		margin: '19px 0 0 0',
	} as React.CSSProperties,
	button: {
		margin: '8px 0 0 8px',
	} as React.CSSProperties,
	buttons: {
		marginTop: '8px',
	} as React.CSSProperties,
	group: {
		flex: 1,
		minWidth: '250px',
		margin: '0 10px',
	} as React.CSSProperties,
	chartGroup: {
		flex: 1,
		minWidth: '250px',
		margin: '0 10px',
		marginBottom: '15px',
	} as React.CSSProperties,
	componentsHeader: {
		fontSize: '20px',
		marginTop: '10px',
		paddingBottom: '2px',
		marginBottom: '10px',
		borderBottomStyle: 'solid',
	} as React.CSSProperties,
	componentsSearch: {
		margin: '12px 0 0 8px',
		width: '220px',
	} as React.CSSProperties,
	componentGroup: {
		flex: 1,
		minWidth: '250px',
		margin: '0 10px 15px 10px',
	} as React.CSSProperties,
	componentLabel: {
		fontSize: '14px',
		fontWeight: 'bold',
		margin: '0 0 6px 0',
	} as React.CSSProperties,
	componentCount: {
		fontWeight: 'normal',
		marginLeft: '6px',
	} as React.CSSProperties,
	componentList: {
		display: 'flex',
		flexWrap: 'wrap',
		alignItems: 'flex-start',
		maxHeight: '260px',
		overflowY: 'auto',
		padding: '4px 6px 8px 6px',
		borderRadius: '3px',
	} as React.CSSProperties,
	componentTag: {
		margin: '4px 3px 0 3px',
		minHeight: '20px',
		fontFamily: 'monospace',
	} as React.CSSProperties,
	componentEmpty: {
		margin: '8px 6px',
	} as React.CSSProperties,
};

export default class MetricCharts extends React.Component<Props, State> {
	loading: {[key: string]: boolean};
	chartBoxRef: React.RefObject<HTMLDivElement>;

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			sync: 0,
			period: 1440,
			interval: 30,
			loading: {},
			cancelable: {},
			componentsFilter: '',
		};

		this.loading = {};
		this.chartBoxRef = React.createRef();
	}

	componentNames(type: string): string[] {
		let filter = (this.state.componentsFilter || '').toLowerCase();
		let names: string[] = [];

		for (let component of (this.props.components || [])) {
			if (component.type !== type || !component.name) {
				continue;
			}
			if (filter && component.name.toLowerCase().indexOf(filter) === -1) {
				continue;
			}
			names.push(component.name);
		}

		names.sort((a: string, b: string): number => {
			let portA = a.match(/^(\w+)\/(\d+)$/);
			let portB = b.match(/^(\w+)\/(\d+)$/);
			if (portA && portB) {
				if (portA[1] !== portB[1]) {
					return portA[1] < portB[1] ? -1 : 1;
				}
				return parseInt(portA[2], 10) - parseInt(portB[2], 10);
			}
			return a < b ? -1 : (a > b ? 1 : 0);
		});
		return names;
	}

	renderComponentGroup(label: string, type: string,
			icon: string): JSX.Element {

		let names = this.componentNames(type);
		let total = 0;
		for (let component of (this.props.components || [])) {
			if (component.type === type) {
				total += 1;
			}
		}

		let count = names.length.toString();
		if (names.length !== total) {
			count = names.length + ' of ' + total;
		}

		let tags: JSX.Element[] = [];
		for (let name of names) {
			tags.push(
				<div
					className="bp5-tag bp5-minimal"
					style={css.componentTag}
					key={name}
				>
					{name}
				</div>,
			);
		}

		let body: JSX.Element;
		if (tags.length) {
			body = <div className="bp5-card bp5-elevation-0" style={css.componentList}>
				{tags}
			</div>;
		} else {
			body = <div className="bp5-card bp5-elevation-0" style={css.componentList}>
				<span className="bp5-text-muted" style={css.componentEmpty}>
					{total ? 'No matches' : 'None reported'}
				</span>
			</div>;
		}

		return <div style={css.componentGroup}>
			<h5 style={css.componentLabel}>
				<span className={'bp5-icon-standard bp5-icon-' + icon}/> {label}
				<span className="bp5-text-muted" style={css.componentCount}>
					({count})
				</span>
			</h5>
			{body}
		</div>;
	}

	renderComponents(): JSX.Element {
		if (!this.props.instance && !this.props.node) {
			return null;
		}

		return <div>
			<div
				className="layout horizontal wrap bp5-border"
				style={css.componentsHeader}
			>
				<h3 style={css.heading}>Components</h3>
				<div className="flex"/>
				<SearchInput
					style={css.componentsSearch}
					placeholder="Filter components"
					value={this.state.componentsFilter}
					onChange={(val: string): void => {
						this.setState({
							...this.state,
							componentsFilter: val,
						});
					}}
				/>
			</div>
			<div className="layout horizontal wrap">
				{this.renderComponentGroup('Processes', 'process', 'application')}
				{this.renderComponentGroup('Kernel Modules', 'module', 'cog')}
				{this.renderComponentGroup('Ports', 'port', 'globe-network')}
			</div>
		</div>;
	}

	getDefaultInterval(period: number): number {
		switch (period) {
			case 60:
				return 1;
			case 180:
				return 5;
			case 360:
				return 5;
			case 720:
				return 30;
			case 1440:
				return 30;
			case 4320:
				return 60;
			case 10080:
				return 120;
			case 20160:
				return 360;
			case 43200:
				return 720;
			case 86400:
				return 1440;
			case 129600:
				return 1440;
			case 172800:
				return 4320;
			default:
				return 360;
		}
	}

	setLoading(resource: string): void {
		this.loading[resource] = true;

		let loading = {
			...this.state.loading,
		};
		loading[resource] = true;

		setTimeout((): void => {
			if (this.loading[resource]) {
				let cancelable = {
					...this.state.cancelable,
				};
				cancelable[resource] = true;

				this.setState({
					...this.state,
					cancelable: cancelable,
				});
			}
		}, 3000);

		this.setState({
			...this.state,
			loading: loading,
		});
	}

	setLoaded(resource: string): void {
		delete this.loading[resource];

		let loading = {
			...this.state.loading,
		};
		delete loading[resource];

		let cancelable = {
			...this.state.cancelable,
		};
		delete cancelable[resource];

		this.setState({
			...this.state,
			loading: loading,
			cancelable: cancelable,
		});
	}

	render(): JSX.Element {
		if (this.props.disabled) {
			return <div/>;
		}

		let intervalMin = 0;
		let intervalMax = 0;
		if (this.state.period > 43200) {
			intervalMin = 120;
		} else if (this.state.period > 20160) {
			intervalMin = 30;
		} else if (this.state.period > 4320) {
			intervalMin = 5;
		}

		if (this.state.period <= 60) {
			intervalMax = 30;
		} else if (this.state.period <= 180) {
			intervalMax = 60;
		} else if (this.state.period <= 360) {
			intervalMax = 120;
		} else if (this.state.period <= 720) {
			intervalMax = 360;
		} else if (this.state.period <= 1440) {
			intervalMax = 720;
		} else if (this.state.period <= 4320) {
			intervalMax = 1440;
		} else if (this.state.period <= 10080) {
			intervalMax = 4320;
		} else {
			intervalMax = 10080;
		}

		let refreshDisabled = false;
		let refreshLabel = '';
		let refreshClass = 'bp5-button';
		if (Object.entries(this.state.cancelable).length) {
			refreshLabel = 'Cancel';
			refreshClass += ' bp5-intent-warning bp5-icon-delete'
		} else {
			if (Object.entries(this.state.loading).length) {
				refreshDisabled = true;
			}
			refreshLabel = 'Refresh';
			refreshClass += ' bp5-intent-success bp5-icon-refresh'
		}

		return <div ref={this.chartBoxRef}>
			<div className="layout horizontal wrap bp5-border" style={css.header}>
				<h3 style={css.heading}>Charts</h3>
				<div className="flex"/>
				<div style={css.buttons}>
					<button
						className={refreshClass}
						style={css.button}
						disabled={refreshDisabled}
						type="button"
						onClick={(): void => {
							if (Object.entries(this.state.cancelable).length) {
								if (this.props.node) {
									NodeActions.dataCancel();
								} else {
									InstanceActions.dataCancel();
								}
							} else {
								this.setState({
									...this.state,
									sync: this.state.sync + 1,
								});
							}
						}}
					>
						{refreshLabel}
					</button>
				</div>
			</div>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<PageSelect
						label="Time Range"
						help="Select chart time range."
						value={this.state.period.toString()}
						onChange={(val: string): void => {
							let period = parseInt(val, 10);
							this.setState({
								...this.state,
								period: period,
								interval: this.getDefaultInterval(period),
							});
						}}
					>
						<option value="60">1 hour</option>
						<option value="180">3 hours</option>
						<option value="360">6 hours</option>
						<option value="720">12 hours</option>
						<option value="1440">24 hours</option>
						<option value="4320">3 days</option>
						<option value="10080">7 days</option>
						<option value="20160">14 days</option>
						<option value="43200">30 days</option>
						<option value="86400">60 days</option>
						<option value="129600" hidden={true}>90 days</option>
						<option value="172800" hidden={true}>120 days</option>
					</PageSelect>
				</div>
				<div style={css.group}>
					<PageSelect
						label="Interval"
						help="Select chart interval."
						value={this.state.interval.toString()}
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								interval: parseInt(val, 10),
							});
						}}
					>
						<option
							value="1"
							hidden={1 < intervalMin || 1 > intervalMax}
						>1 Minute</option>
						<option
							value="5"
							hidden={5 < intervalMin || 5 > intervalMax}
						>5 Minutes</option>
						<option
							value="30"
							hidden={30 < intervalMin || 30 > intervalMax}
						>30 Minutes</option>
						<option
							value="60"
							hidden={60 < intervalMin || 60 > intervalMax}
						>1 Hour</option>
						<option
							value="120"
							hidden={120 < intervalMin || 120 > intervalMax}
						>2 Hours</option>
						<option
							value="360"
							hidden={360 < intervalMin || 360 > intervalMax}
						>6 Hours</option>
						<option
							value="720"
							hidden={720 < intervalMin || 720 > intervalMax}
						>12 Hours</option>
						<option
							value="1440"
							hidden={1440 < intervalMin || 1440 > intervalMax}
						>24 Hours</option>
						<option
							value="4320"
							hidden={4320 < intervalMin || 4320 > intervalMax}
						>3 Days</option>
						<option
							value="10080"
							hidden={10080 < intervalMin || 10080 > intervalMax}
						>7 Days</option>
					</PageSelect>
				</div>
			</div>
			<div className="layout horizontal wrap">
				<div style={css.chartGroup}>
					<MetricChart
						instance={this.props.instance}
						node={this.props.node}
						resource={'system'}
						sync={this.state.sync}
						period={this.state.period}
						interval={this.state.interval}
						left={true}
						onLoading={(): void => {
							this.setLoading('system');
						}}
						onLoaded={(): void => {
							this.setLoaded('system');
						}}
						getBoxRect={(): DOMRect => {
							return this.chartBoxRef.current.getBoundingClientRect();
						}}
					/>
				</div>
				<div style={css.chartGroup}>
					<MetricChart
						instance={this.props.instance}
						node={this.props.node}
						resource={'load'}
						sync={this.state.sync}
						period={this.state.period}
						interval={this.state.interval}
						left={false}
						onLoading={(): void => {
							this.setLoading('load');
						}}
						onLoaded={(): void => {
							this.setLoaded('load');
						}}
						getBoxRect={(): DOMRect => {
							return this.chartBoxRef.current.getBoundingClientRect();
						}}
					/>
				</div>
			</div>
			<div className="layout horizontal wrap">
				<div style={css.chartGroup}>
					<MetricChart
						instance={this.props.instance}
						node={this.props.node}
						resource={'disk'}
						sync={this.state.sync}
						period={this.state.period}
						interval={this.state.interval}
						left={true}
						onLoading={(): void => {
							this.setLoading('disk');
						}}
						onLoaded={(): void => {
							this.setLoaded('disk');
						}}
						getBoxRect={(): DOMRect => {
							return this.chartBoxRef.current.getBoundingClientRect();
						}}
					/>
				</div>
				<div style={css.chartGroup}>
					<MetricChart
						instance={this.props.instance}
						node={this.props.node}
						resource={'network'}
						sync={this.state.sync}
						period={this.state.period}
						interval={this.state.interval}
						left={false}
						onLoading={(): void => {
							this.setLoading('network');
						}}
						onLoaded={(): void => {
							this.setLoaded('network');
						}}
						getBoxRect={(): DOMRect => {
							return this.chartBoxRef.current.getBoundingClientRect();
						}}
					/>
				</div>
			</div>
			<div className="layout horizontal wrap">
				<div style={css.chartGroup}>
					<MetricChart
						instance={this.props.instance}
						node={this.props.node}
						resource={'diskio0'}
						sync={this.state.sync}
						period={this.state.period}
						interval={this.state.interval}
						left={true}
						onLoading={(): void => {
							this.setLoading('diskio0');
						}}
						onLoaded={(): void => {
							this.setLoaded('diskio0');
						}}
						getBoxRect={(): DOMRect => {
							return this.chartBoxRef.current.getBoundingClientRect();
						}}
					/>
				</div>
				<div style={css.chartGroup}>
					<MetricChart
						instance={this.props.instance}
						node={this.props.node}
						resource={'diskio1'}
						sync={this.state.sync}
						period={this.state.period}
						interval={this.state.interval}
						left={false}
						onLoading={(): void => {
							this.setLoading('diskio1');
						}}
						onLoaded={(): void => {
							this.setLoaded('diskio1');
						}}
						getBoxRect={(): DOMRect => {
							return this.chartBoxRef.current.getBoundingClientRect();
						}}
					/>
				</div>
			</div>
			{this.renderComponents()}
		</div>;
	}
}
