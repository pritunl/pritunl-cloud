/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as InstanceTypes from '../types/InstanceTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as NodeTypes from '../types/NodeTypes';
import * as ImageTypes from '../types/ImageTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import * as InstanceActions from '../actions/InstanceActions';
import * as ImageActions from '../actions/ImageActions';
import * as NodeActions from '../actions/NodeActions';
import ImagesDatacenterStore from '../stores/ImagesDatacenterStore';
import NodesZoneStore from '../stores/NodesZoneStore';
import PageInput from './PageInput';
import PageCreate from './PageCreate';
import PageSelect from './PageSelect';
import PageNumInput from './PageNumInput';

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	instance: InstanceTypes.Instance;
	datacenter: string;
	images: ImageTypes.ImagesRo;
	nodes: NodeTypes.NodesRo;
}

const css = {
	row: {
		display: 'table-row',
		width: '100%',
		padding: 0,
		boxShadow: 'none',
		position: 'relative',
	} as React.CSSProperties,
	card: {
		position: 'relative',
		padding: '10px 10px 0 10px',
		width: '100%',
	} as React.CSSProperties,
	buttons: {
		position: 'absolute',
		top: '5px',
		right: '5px',
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	itemsLabel: {
		display: 'block',
	} as React.CSSProperties,
	itemsAdd: {
		margin: '8px 0 15px 0',
	} as React.CSSProperties,
	group: {
		flex: 1,
		minWidth: '250px',
	} as React.CSSProperties,
	save: {
		paddingBottom: '10px',
	} as React.CSSProperties,
	label: {
		width: '100%',
		maxWidth: '280px',
	} as React.CSSProperties,
	inputGroup: {
		width: '100%',
	} as React.CSSProperties,
	protocol: {
		flex: '0 1 auto',
	} as React.CSSProperties,
	port: {
		flex: '1',
	} as React.CSSProperties,
};

export default class InstanceNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			instance: this.default,
			datacenter: '',
			images: [],
			nodes: [],
		};
	}

	componentDidMount(): void {
		ImagesDatacenterStore.addChangeListener(this.onChange);
		NodesZoneStore.addChangeListener(this.onChange);
	}

	componentWillUnmount(): void {
		ImagesDatacenterStore.removeChangeListener(this.onChange);
		NodesZoneStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			images: ImagesDatacenterStore.images,
			nodes: NodesZoneStore.nodes,
		});
	}

	get default(): InstanceTypes.Instance {
		return {
			id: null,
			name: 'New instance',
			memory: 1024,
			processors: 1,
			count: 1,
		};
	}

	set(name: string, val: any): void {
		let instance: any = {
			...this.state.instance,
		};

		instance[name] = val;

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let instance: any = {
			...this.state.instance,
		};

		if (this.props.organizations.length && !instance.organization) {
			instance.organization = this.props.organizations[0].id;
		}

		InstanceActions.create(instance).then((): void => {
			this.setState({
				...this.state,
				message: 'Instance created successfully',
				closed: true,
				changed: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let instance = this.state.instance;

		let organizationsSelect: JSX.Element[] = [];
		if (this.props.organizations.length) {
			for (let organization of this.props.organizations) {
				organizationsSelect.push(
					<option
						key={organization.id}
						value={organization.id}
					>{organization.name}</option>,
				);
			}
		}

		let defaultDatacenter = '';
		let hasDatacenters = false;
		let datacentersSelect: JSX.Element[] = [];
		if (this.props.datacenters.length) {
			datacentersSelect.push(
				<option key="null" value="">Select Datacenter</option>);

			hasDatacenters = true;
			defaultDatacenter = this.props.datacenters[0].id;
			for (let datacenter of this.props.datacenters) {
				datacentersSelect.push(
					<option
						key={datacenter.id}
						value={datacenter.id}
					>{datacenter.name}</option>,
				);
			}
		}

		if (!hasDatacenters) {
			datacentersSelect.push(
				<option key="null" value="">No Datacenters</option>);
		}

		let datacenter = this.state.datacenter || defaultDatacenter;
		let hasZones = false;
		let zonesSelect: JSX.Element[] = [];
		if (this.props.zones.length) {
			zonesSelect.push(<option key="null" value="">Select Zone</option>);

			for (let zone of this.props.zones) {
				if (zone.datacenter !== datacenter) {
					continue;
				}
				hasZones = true;

				zonesSelect.push(
					<option
						key={zone.id}
						value={zone.id}
					>{zone.name}</option>,
				);
			}
		}

		if (!hasZones) {
			zonesSelect = [<option key="null" value="">No Zones</option>];
		}

		let hasNodes = false;
		let nodesSelect: JSX.Element[] = [];
		if (this.state.nodes.length) {
			nodesSelect.push(<option key="null" value="">Select Node</option>);

			hasNodes = true;
			for (let node of this.state.nodes) {
				nodesSelect.push(
					<option
						key={node.id}
						value={node.id}
					>{node.name}</option>,
				);
			}
		}

		if (!hasNodes) {
			nodesSelect = [<option key="null" value="">No Nodes</option>];
		}

		let hasImages = false;
		let imagesSelect: JSX.Element[] = [];
		if (this.state.images.length) {
			imagesSelect.push(<option key="null" value="">Select Image</option>);

			hasImages = true;
			for (let image of this.state.images) {
				imagesSelect.push(
					<option
						key={image.id}
						value={image.id}
					>{image.name}</option>,
				);
			}
		}

		if (!hasImages) {
			imagesSelect = [<option key="null" value="">No Images</option>];
		}

		return <div
			className="pt-card pt-row"
			style={css.row}
		>
			<td
				className="pt-cell"
				colSpan={5}
				style={css.card}
			>
				<div className="layout horizontal wrap">
					<div style={css.group}>
						<div style={css.buttons}>
						</div>
						<PageInput
							label="Name"
							help="Name of instance"
							type="text"
							placeholder="Enter name"
							disabled={this.state.disabled}
							value={instance.name}
							onChange={(val): void => {
								this.set('name', val);
							}}
						/>
						<PageSelect
							disabled={this.state.disabled}
							label="Organization"
							help="Organization for instance."
							value={instance.organization}
							onChange={(val): void => {
								this.set('organization', val);
							}}
						>
							{organizationsSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasDatacenters}
							label="Datacenter"
							help="Datacenter for instance."
							value={this.state.datacenter}
							onChange={(val): void => {
								this.setState({
									...this.state,
									datacenter: val,
									instance: {
										...this.state.instance,
										node: '',
										zone: '',
										image: '',
									},
								});
								ImageActions.syncDatacenter(val);
								NodeActions.syncZone('');
							}}
						>
							{datacentersSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasZones}
							label="Zone"
							help="Zone for instance."
							value={instance.zone}
							onChange={(val): void => {
								this.setState({
									...this.state,
									instance: {
										...this.state.instance,
										node: '',
										zone: val,
									},
								});
								NodeActions.syncZone(val);
							}}
						>
							{zonesSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasNodes}
							label="Node"
							help="Node to run instance on."
							value={instance.node}
							onChange={(val): void => {
								this.set('node', val);
							}}
						>
							{nodesSelect}
						</PageSelect>
					</div>
					<div style={css.group}>
						<PageSelect
							disabled={this.state.disabled || !hasImages}
							label="Image"
							help="Starting image for node."
							value={instance.image}
							onChange={(val): void => {
								this.set('image', val);
							}}
						>
							{imagesSelect}
						</PageSelect>
						<PageNumInput
							label="Memory Size"
							help="Instance memory size in megabytes."
							min={256}
							minorStepSize={256}
							stepSize={512}
							majorStepSize={1024}
							disabled={this.state.disabled}
							selectAllOnFocus={true}
							onChange={(val: number): void => {
								this.set('memory', val);
							}}
							value={instance.memory}
						/>
						<PageNumInput
							label="Processors"
							help="Number of instance processors."
							min={1}
							minorStepSize={1}
							stepSize={1}
							majorStepSize={2}
							disabled={this.state.disabled}
							selectAllOnFocus={true}
							onChange={(val: number): void => {
								this.set('processors', val);
							}}
							value={instance.processors}
						/>
						<PageNumInput
							label="Count"
							help="Number of instances to create."
							min={1}
							minorStepSize={1}
							stepSize={1}
							majorStepSize={1}
							disabled={this.state.disabled}
							selectAllOnFocus={true}
							onChange={(val: number): void => {
								this.set('count', val);
							}}
							value={instance.count}
						/>
					</div>
				</div>
				<PageCreate
					style={css.save}
					hidden={!this.state.instance}
					message={this.state.message}
					changed={this.state.changed}
					disabled={this.state.disabled}
					closed={this.state.closed}
					light={true}
					onCancel={this.props.onClose}
					onCreate={this.onCreate}
				/>
			</td>
		</div>;
	}
}
