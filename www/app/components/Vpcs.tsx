/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as VpcTypes from '../types/VpcTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import VpcsStore from '../stores/VpcsStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import * as VpcActions from '../actions/VpcActions';
import Vpc from './Vpc';
import VpcNew from './VpcNew';
import VpcsFilter from './VpcsFilter';
import VpcsPage from './VpcsPage';
import Page from './Page';
import PageHeader from './PageHeader';
import NonState from './NonState';
import ConfirmButton from './ConfirmButton';
import CompletionStore from "../stores/CompletionStore";
import * as CompletionActions from "../actions/CompletionActions";
import * as DatacenterTypes from "../types/DatacenterTypes";

interface Selected {
	[key: string]: boolean;
}

interface Opened {
	[key: string]: boolean;
}

interface State {
	vpcs: VpcTypes.VpcsRo;
	filter: VpcTypes.Filter;
	datacenters: DatacenterTypes.DatacentersRo;
	organizations: OrganizationTypes.OrganizationsRo;
	network: string;
	datacenter: string;
	organization: string;
	selected: Selected;
	opened: Opened;
	newOpened: boolean;
	lastSelected: string;
	disabled: boolean;
}

const css = {
	items: {
		width: '100%',
		marginTop: '-5px',
		display: 'table',
		tableLayout: 'fixed',
		borderSpacing: '0 5px',
	} as React.CSSProperties,
	itemsBox: {
		width: '100%',
		overflowY: 'auto',
	} as React.CSSProperties,
	group: {
		width: '100%',
	} as React.CSSProperties,
	groupBox: {
		margin: '16px 0 0 0',
		width: '100%',
		maxWidth: '420px',
	} as React.CSSProperties,
	groupBoxUser: {
		margin: '16px 0 0 0',
		width: '100%',
		maxWidth: '310px',
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
	input: {
		width: '107px',
	} as React.CSSProperties,
	select: {
		width: '100%',
	} as React.CSSProperties,
	selectFirst: {
		width: '100%',
		borderTopLeftRadius: '3px',
		borderBottomLeftRadius: '3px',
	} as React.CSSProperties,
	selectInner: {
		width: '100%',
	} as React.CSSProperties,
	selectBox: {
		flex: '1',
	} as React.CSSProperties,
	button: {
		margin: '8px 0 0 8px',
	} as React.CSSProperties,
	buttons: {
		margin: '8px 8px 0 0',
	} as React.CSSProperties,
};

export default class Vpcs extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			vpcs: VpcsStore.vpcs,
			filter: VpcsStore.filter,
			datacenters: CompletionStore.datacenters,
			organizations: CompletionStore.organizations,
			network: '',
			organization: '',
			datacenter: '',
			selected: {},
			opened: {},
			newOpened: false,
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
		VpcsStore.addChangeListener(this.onChange);
		CompletionStore.addChangeListener(this.onChange);
		OrganizationsStore.addChangeListener(this.onChange);
		VpcActions.sync();
		CompletionActions.sync();
	}

	componentWillUnmount(): void {
		VpcsStore.removeChangeListener(this.onChange);
		CompletionStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let vpcs = VpcsStore.vpcs;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		vpcs.forEach((vpc: VpcTypes.Vpc): void => {
			if (curSelected[vpc.id]) {
				selected[vpc.id] = true;
			}
			if (curOpened[vpc.id]) {
				opened[vpc.id] = true;
			}
		});

		this.setState({
			...this.state,
			vpcs: vpcs,
			filter: VpcsStore.filter,
			datacenters: CompletionStore.datacenters,
			organizations: CompletionStore.organizations,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		VpcActions.removeMulti(
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
		let vpcsDom: JSX.Element[] = [];

		let hasOrganizations = false;
		let organizationsSelect: JSX.Element[] = [];
		if (this.state.organizations.length) {
			hasOrganizations = true;
			for (let organization of this.state.organizations) {
				organizationsSelect.push(
					<option
						key={organization.id}
						value={organization.id}
					>{organization.name}</option>,
				);
			}
		} else {
			organizationsSelect.push(
				<option
					key="null"
					value=""
				>No Organizations</option>,
			);
		}

		let hasDatacenters = false;
		let datacentersSelect: JSX.Element[] = [];
		if (this.state.datacenters.length) {
			hasDatacenters = true;
			for (let datacenter of this.state.datacenters) {
				datacentersSelect.push(
					<option
						key={datacenter.id}
						value={datacenter.id}
					>{datacenter.name}</option>,
				);
			}
		} else {
			datacentersSelect.push(
				<option
					key="null"
					value=""
				>No Datacenters</option>,
			);
		}

		this.state.vpcs.forEach((
				vpc: VpcTypes.VpcRo): void => {
			vpcsDom.push(<Vpc
				key={vpc.id}
				vpc={vpc}
				organizations={this.state.organizations}
				selected={!!this.state.selected[vpc.id]}
				open={!!this.state.opened[vpc.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let vpcs = this.state.vpcs;
						let start: number;
						let end: number;

						for (let i = 0; i < vpcs.length; i++) {
							let usr = vpcs[i];

							if (usr.id === vpc.id) {
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
								selected[vpcs[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: vpc.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[vpc.id]) {
						delete selected[vpc.id];
					} else {
						selected[vpc.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: vpc.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[vpc.id]) {
						delete opened[vpc.id];
					} else {
						opened[vpc.id] = true;
					}

					this.setState({
						...this.state,
						opened: opened,
					});
				}}
			/>);
		});

		let filterClass = 'bp5-button bp5-intent-primary bp5-icon-filter ';
		if (this.state.filter) {
			filterClass += 'bp5-active';
		}

		let selectedNames: string[] = [];
		for (let vpcId of Object.keys(this.state.selected)) {
			let inst = VpcsStore.vpc(vpcId);
			if (inst) {
				selectedNames.push(inst.name || vpcId);
			} else {
				selectedNames.push(vpcId);
			}
		}

		let newVpcDom: JSX.Element;
		if (this.state.newOpened) {
			newVpcDom = <VpcNew
				organizations={this.state.organizations}
				datacenters={this.state.datacenters}
				onClose={(): void => {
					this.setState({
						...this.state,
						newOpened: false,
					});
				}}
			/>;
		}

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>VPCs</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									VpcActions.filter({});
								} else {
									VpcActions.filter(null);
								}
							}}
						>
							Filters
						</button>
						<button
							className="bp5-button bp5-intent-warning bp5-icon-chevron-up"
							style={css.button}
							disabled={!this.opened}
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
							className="bp5-intent-danger bp5-icon-delete"
							progressClassName="bp5-intent-danger"
							safe={true}
							style={css.button}
							confirmMsg="Permanently delete the selected VPCs"
							confirmInput={true}
							items={selectedNames}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onDelete}
						/>
						<button
							className="bp5-button bp5-intent-success bp5-icon-add"
							style={css.button}
							disabled={this.state.disabled || this.state.newOpened}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									newOpened: true,
								});
							}}
						>New</button>
					</div>
				</div>
			</PageHeader>
			<VpcsFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					VpcActions.filter(filter);
				}}
				organizations={this.state.organizations}
				datacenters={this.state.datacenters}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newVpcDom}
					{vpcsDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={6} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!vpcsDom.length}
				iconClass="bp5-icon-layout-auto"
				title="No vpcs"
				description="Add a new vpc to get started."
			/>
			<VpcsPage
				onPage={(): void => {
					this.setState({
						...this.state,
						selected: {},
						lastSelected: null,
					});
				}}
			/>
		</Page>;
	}
}
