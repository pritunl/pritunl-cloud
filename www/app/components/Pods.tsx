/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as PodTypes from '../types/PodTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import PodsStore from '../stores/PodsStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import * as PodActions from '../actions/PodActions';
import * as OrganizationActions from '../actions/OrganizationActions';
import Pod from './Pod';
import PodNew from './PodNew';
import PodsFilter from './PodsFilter';
import PodsPage from './PodsPage';
import Page from './Page';
import PageHeader from './PageHeader';
import NonState from './NonState';
import ConfirmButton from './ConfirmButton';
import DomainsNameStore from "../stores/DomainsNameStore";
import VpcsNameStore from "../stores/VpcsNameStore";
import DatacentersStore from "../stores/DatacentersStore";
import NodesStore from "../stores/NodesStore";
import PoolsStore from "../stores/PoolsStore";
import ZonesStore from "../stores/ZonesStore";
import ShapesStore from "../stores/ShapesStore";
import InstancesStore from "../stores/InstancesStore";
import ImagesStore from "../stores/ImagesStore";
import PlansStore from "../stores/PlansStore";
import CertificatesStore from "../stores/CertificatesStore";
import SecretsStore from "../stores/SecretsStore";
import CompletionStore from "../completion/Store"
import * as VpcTypes from "../types/VpcTypes";
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as NodeTypes from "../types/NodeTypes";
import * as PoolTypes from "../types/PoolTypes";
import * as ZoneTypes from "../types/ZoneTypes";
import * as ShapeTypes from "../types/ShapeTypes";
import * as ImageTypes from "../types/ImageTypes";
import * as InstanceTypes from "../types/InstanceTypes";
import * as PlanTypes from "../types/PlanTypes";
import * as CertificateTypes from "../types/CertificateTypes";
import * as SecretTypes from "../types/SecretTypes";
import * as DomainActions from "../actions/DomainActions";
import * as VpcActions from "../actions/VpcActions";
import * as DatacenterActions from "../actions/DatacenterActions";
import * as NodeActions from "../actions/NodeActions";
import * as PoolActions from "../actions/PoolActions";
import * as ZoneActions from "../actions/ZoneActions";
import * as ShapeActions from "../actions/ShapeActions";
import * as ImageActions from "../actions/ImageActions";
import * as InstanceActions from "../actions/InstanceActions";
import * as PlanActions from "../actions/PlanActions";
import * as CertificateActions from "../actions/CertificateActions";
import * as SecretActions from "../actions/SecretActions";
import * as DomainTypes from "../types/DomainTypes";

interface Selected {
	[key: string]: boolean;
}

interface Opened {
	[key: string]: boolean;
}

interface State {
	pods: PodTypes.PodsRo;
	filter: PodTypes.Filter;
	organizations: OrganizationTypes.OrganizationsRo;
	domains: DomainTypes.DomainsRo;
	vpcs: VpcTypes.VpcsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	nodes: NodeTypes.NodesRo;
	pools: PoolTypes.PoolsRo;
	zones: ZoneTypes.ZonesRo;
	shapes: ShapeTypes.ShapesRo;
	images: ImageTypes.ImagesRo;
	instances: InstanceTypes.InstancesRo;
	plans: PlanTypes.PlansRo;
	certificates: CertificateTypes.CertificatesRo;
	secrets: SecretTypes.SecretsRo;
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
		margin: '8px 0 0 8px',
	} as React.CSSProperties,
	buttons: {
		marginTop: '8px',
	} as React.CSSProperties,
};

export default class Pods extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			pods: PodsStore.pods,
			filter: PodsStore.filter,
			organizations: OrganizationsStore.organizations,
			domains: DomainsNameStore.domains,
			vpcs: VpcsNameStore.vpcs,
			datacenters: DatacentersStore.datacenters,
			nodes: NodesStore.nodes,
			pools: PoolsStore.pools,
			zones: ZonesStore.zones,
			shapes: ShapesStore.shapes,
			images: ImagesStore.images,
			instances: InstancesStore.instances,
			plans: PlansStore.plans,
			certificates: CertificatesStore.certificates,
			secrets: SecretsStore.secrets,
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
		PodsStore.addChangeListener(this.onChange);
		OrganizationsStore.addChangeListener(this.onChange);
		DomainsNameStore.addChangeListener(this.onChange);
		VpcsNameStore.addChangeListener(this.onChange);
		DatacentersStore.addChangeListener(this.onChange);
		NodesStore.addChangeListener(this.onChange);
		PoolsStore.addChangeListener(this.onChange);
		ZonesStore.addChangeListener(this.onChange);
		ShapesStore.addChangeListener(this.onChange);
		ImagesStore.addChangeListener(this.onChange);
		InstancesStore.addChangeListener(this.onChange);
		PlansStore.addChangeListener(this.onChange);
		CertificatesStore.addChangeListener(this.onChange);
		SecretsStore.addChangeListener(this.onChange);
		PodActions.sync();
		OrganizationActions.sync();
		DomainActions.syncName();
		VpcActions.syncNames();
		DatacenterActions.sync();
		NodeActions.sync();
		PoolActions.sync();
		ZoneActions.sync();
		ShapeActions.sync();
		ImageActions.sync();
		InstanceActions.sync();
		PlanActions.sync();
		CertificateActions.sync();
		SecretActions.sync();
	}

	componentWillUnmount(): void {
		PodsStore.removeChangeListener(this.onChange);
		OrganizationsStore.removeChangeListener(this.onChange);
		DomainsNameStore.removeChangeListener(this.onChange);
		VpcsNameStore.removeChangeListener(this.onChange);
		DatacentersStore.removeChangeListener(this.onChange);
		NodesStore.removeChangeListener(this.onChange);
		PoolsStore.removeChangeListener(this.onChange);
		ZonesStore.removeChangeListener(this.onChange);
		ShapesStore.removeChangeListener(this.onChange);
		ImagesStore.removeChangeListener(this.onChange);
		InstancesStore.removeChangeListener(this.onChange);
		PlansStore.removeChangeListener(this.onChange);
		CertificatesStore.removeChangeListener(this.onChange);
		SecretsStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let pods = PodsStore.pods;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;
		let units: PodTypes.Units = []

		pods.forEach((pod: PodTypes.Pod): void => {
			pod.units.forEach((unit: PodTypes.Unit): void => {
				units.push(unit)
			})

			if (curSelected[pod.id]) {
				selected[pod.id] = true;
			}
			if (curOpened[pod.id]) {
				opened[pod.id] = true;
			}
		});

		CompletionStore.update({
			organizations: OrganizationsStore.organizations,
			domains: DomainsNameStore.domains,
			vpcs: VpcsNameStore.vpcs,
			datacenters: DatacentersStore.datacenters,
			nodes: NodesStore.nodes,
			pools: PoolsStore.pools,
			zones: ZonesStore.zones,
			shapes: ShapesStore.shapes,
			images: ImagesStore.images,
			instances: InstancesStore.instances,
			plans: PlansStore.plans,
			certificates: CertificatesStore.certificates,
			secrets: SecretsStore.secrets,
			pods: PodsStore.pods,
			units: units,
		})

		this.setState({
			...this.state,
			pods: pods,
			filter: PodsStore.filter,
			organizations: OrganizationsStore.organizations,
			domains: DomainsNameStore.domains,
			vpcs: VpcsNameStore.vpcs,
			datacenters: DatacentersStore.datacenters,
			nodes: NodesStore.nodes,
			pools: PoolsStore.pools,
			zones: ZonesStore.zones,
			shapes: ShapesStore.shapes,
			images: ImagesStore.images,
			instances: InstancesStore.instances,
			plans: PlansStore.plans,
			certificates: CertificatesStore.certificates,
			secrets: SecretsStore.secrets,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		PodActions.removeMulti(
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
		let podsDom: JSX.Element[] = [];

		this.state.pods.forEach((
				pod: PodTypes.PodRo): void => {
			podsDom.push(<Pod
				key={pod.id}
				pod={pod}
				organizations={this.state.organizations}
				selected={!!this.state.selected[pod.id]}
				open={!!this.state.opened[pod.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let pods = this.state.pods;
						let start: number;
						let end: number;

						for (let i = 0; i < pods.length; i++) {
							let usr = pods[i];

							if (usr.id === pod.id) {
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
								selected[pods[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: pod.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[pod.id]) {
						delete selected[pod.id];
					} else {
						selected[pod.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: pod.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[pod.id]) {
						delete opened[pod.id];
					} else {
						opened[pod.id] = true;
					}

					this.setState({
						...this.state,
						opened: opened,
					});
				}}
			/>);
		});

		let newPodDom: JSX.Element;
		if (this.state.newOpened) {
			newPodDom = <PodNew
				organizations={this.state.organizations}
				onClose={(): void => {
					this.setState({
						...this.state,
						newOpened: false,
					});
				}}
			/>;
		}

		let filterClass = 'bp5-button bp5-intent-primary bp5-icon-filter ';
		if (this.state.filter) {
			filterClass += 'bp5-active';
		}

		let selectedNames: string[] = [];
		for (let instId of Object.keys(this.state.selected)) {
			let inst = PodsStore.pod(instId);
			if (inst) {
				selectedNames.push(inst.name || instId);
			} else {
				selectedNames.push(instId);
			}
		}

		return <Page wide={true}>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Pods</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									PodActions.filter({});
								} else {
									PodActions.filter(null);
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
							confirmMsg="Permanently delete the selected pods"
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
			<PodsFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					PodActions.filter(filter);
				}}
				organizations={this.state.organizations}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newPodDom}
					{podsDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={5} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!podsDom.length}
				iconClass="bp5-icon-cube"
				title="No pods"
				description="Add a new pod to get started."
			/>
			<PodsPage
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
