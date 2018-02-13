/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as InstanceTypes from '../types/InstanceTypes';
import * as CertificateTypes from '../types/CertificateTypes';
import InstancesStore from '../stores/InstancesStore';
import CertificatesStore from '../stores/CertificatesStore';
import * as InstanceActions from '../actions/InstanceActions';
import * as CertificateActions from '../actions/CertificateActions';
import Instance from './Instance';
import InstancesPage from './InstancesPage';
import Page from './Page';
import PageHeader from './PageHeader';
import NonState from './NonState';
import ConfirmButton from './ConfirmButton';

interface Selected {
	[key: string]: boolean;
}

interface Opened {
	[key: string]: boolean;
}

interface State {
	instances: InstanceTypes.InstancesRo;
	certificates: CertificateTypes.CertificatesRo;
	selected: Selected;
	opened: Opened;
	lastSelected: string;
	disabled: boolean;
}

const css = {
	items: {
		width: '100%',
		marginTop: '-5px',
		display: 'table',
		borderSpacing: '0 5px',
	} as React.CSSProperties,
	itemsBox: {
		width: '100%',
		overflowY: 'auto',
	} as React.CSSProperties,
	placeholder: {
		opacity: 0,
		width: '100%',
	} as React.CSSProperties,
	header: {
		marginTop: '-19px',
	} as React.CSSProperties,
	heading: {
		margin: '19px 0 0 0',
	} as React.CSSProperties,
	button: {
		margin: '10px 0 0 10px',
	} as React.CSSProperties,
};

export default class Instances extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			instances: InstancesStore.instances,
			certificates: CertificatesStore.certificates,
			selected: {},
			opened: {},
			lastSelected: null,
			disabled: false,
		};
	}

	get selected(): boolean {
		return !!Object.keys(this.state.selected).length;
	}

	get opened(): boolean {
		return !!Object.keys(this.state.opened).length;
	}

	componentDidMount(): void {
		InstancesStore.addChangeListener(this.onChange);
		CertificatesStore.addChangeListener(this.onChange);
		InstanceActions.sync();
		CertificateActions.sync();
	}

	componentWillUnmount(): void {
		InstancesStore.removeChangeListener(this.onChange);
		CertificatesStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let instances = InstancesStore.instances;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		instances.forEach((instance: InstanceTypes.Instance): void => {
			if (curSelected[instance.id]) {
				selected[instance.id] = true;
			}
			if (curOpened[instance.id]) {
				opened[instance.id] = true;
			}
		});

		this.setState({
			...this.state,
			instances: instances,
			certificates: CertificatesStore.certificates,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		InstanceActions.removeMulti(
				Object.keys(this.state.selected)).then((): void => {
			this.setState({
				...this.state,
				selected: {},
				disabled: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let instancesDom: JSX.Element[] = [];

		this.state.instances.forEach((
				instance: InstanceTypes.InstanceRo): void => {
			instancesDom.push(<Instance
				key={instance.id}
				instance={instance}
				selected={!!this.state.selected[instance.id]}
				open={!!this.state.opened[instance.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let instances = this.state.instances;
						let start: number;
						let end: number;

						for (let i = 0; i < instances.length; i++) {
							let usr = instances[i];

							if (usr.id === instance.id) {
								start = i;
							} else if (usr.id === this.state.lastSelected) {
								end = i;
							}
						}

						if (start !== undefined && end !== undefined) {
							if (start > end) {
								end = [start, start = end][0];
							}

							for (let i = start; i <= end; i++) {
								selected[instances[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: instance.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[instance.id]) {
						delete selected[instance.id];
					} else {
						selected[instance.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: instance.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[instance.id]) {
						delete opened[instance.id];
					} else {
						opened[instance.id] = true;
					}

					this.setState({
						...this.state,
						opened: opened,
					});
				}}
			/>);
		});

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Instances</h2>
					<div className="flex"/>
					<div>
						<button
							className="pt-button pt-intent-warning pt-icon-chevron-up"
							style={css.button}
							hidden={!this.opened}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									opened: {},
								});
							}}
						>
							Collapse All
						</button>
						<ConfirmButton
							label="Delete Selected"
							className="pt-intent-danger pt-icon-delete"
							progressClassName="pt-intent-danger"
							style={css.button}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onDelete}
						/>
						<button
							className="pt-button pt-intent-success pt-icon-add"
							style={css.button}
							disabled={this.state.disabled}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									disabled: true,
								});
								InstanceActions.create(null).then((): void => {
									this.setState({
										...this.state,
										disabled: false,
									});
								}).catch((): void => {
									this.setState({
										...this.state,
										disabled: false,
									});
								});
							}}
						>New</button>
					</div>
				</div>
			</PageHeader>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{instancesDom}
					<tr className="pt-card pt-row" style={css.placeholder}>
						<td colSpan={4} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!instancesDom.length}
				iconClass="pt-icon-dashboard"
				title="No instances"
				description="Add a new instance to get started."
			/>
			<InstancesPage
				onPage={(): void => {
					this.setState({
						lastSelected: null,
					});
				}}
			/>
		</Page>;
	}
}
