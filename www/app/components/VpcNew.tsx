/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as VpcTypes from '../types/VpcTypes';
import * as VpcActions from '../actions/VpcActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as Constants from "../Constants";
import * as PageInfos from './PageInfo';
import DatacentersStore from "../stores/DatacentersStore";
import OrganizationsStore from "../stores/OrganizationsStore";
import VpcRoute from './VpcRoute';
import VpcMap from './VpcMap';
import VpcArp from './VpcArp';
import VpcSubnet from './VpcSubnet';
import PageInput from './PageInput';
import PageSwitch from './PageSwitch';
import PageSelect from './PageSelect';
import PageInfo from './PageInfo';
import PageCreate from './PageCreate';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import PageTextArea from "./PageTextArea";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	addNetworkRole: string;
	addVpc: string;
	vpc: VpcTypes.Vpc;
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
	button: {
		height: '30px',
	} as React.CSSProperties,
	buttons: {
		position: 'absolute',
		top: '5px',
		right: '5px',
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		minHeight: '20px',
	} as React.CSSProperties,
	itemsLabel: {
		display: 'block',
	} as React.CSSProperties,
	itemsAdd: {
		margin: '8px 0 15px 0',
	} as React.CSSProperties,
	group: {
		flex: 1,
		minWidth: '280px',
		margin: '0 10px',
	} as React.CSSProperties,
	save: {
		paddingBottom: '10px',
	} as React.CSSProperties,
	label: {
		width: '100%',
		maxWidth: '280px',
	} as React.CSSProperties,
	status: {
		margin: '6px 0 0 1px',
	} as React.CSSProperties,
	icon: {
		marginRight: '3px',
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
	select: {
		margin: '7px 0px 0px 6px',
		paddingTop: '3px',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		minHeight: '20px',
	} as React.CSSProperties,
	list: {
		marginBottom: '15px',
	} as React.CSSProperties,
};

export default class VpcNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			addNetworkRole: null,
			addVpc: null,
			vpc: {
				name: "New VPC",
			},
		};
	}

	set(name: string, val: any): void {
		let vpc: any = {
			...this.state.vpc,
		};

		vpc[name] = val;

		this.setState({
			...this.state,
			changed: true,
			vpc: vpc,
		});
	}

	onAddSubnet = (i: number): void => {
		let vpc: VpcTypes.Vpc;

		vpc = {
			...this.state.vpc,
		};

		let subnets = [
			...(vpc.subnets || []),
		];

		if (subnets.length === 0) {
			subnets = [{}];
		}

		subnets.splice(i + 1, 0, {} as VpcTypes.Subnet);
		vpc.subnets = subnets;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			vpc: vpc,
		});
	}

	onChangeSubnet(i: number, subnet: VpcTypes.Subnet): void {
		let vpc: VpcTypes.Vpc;

		vpc = {
			...this.state.vpc,
		};

		let subnets = [
			...(vpc.subnets || []),
		];

		if (subnets.length === 0) {
			subnets = [{}];
		}

		subnets[i] = subnet;

		vpc.subnets = subnets;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			vpc: vpc,
		});
	}

	onRemoveSubnet(i: number): void {
		let vpc: VpcTypes.Vpc;

		vpc = {
			...this.state.vpc,
		};

		let subnets = [
			...(vpc.subnets || []),
		];

		if (subnets.length !== 0) {
			subnets.splice(i, 1);
		}

		if (subnets.length === 0) {
			subnets = [{}];
		}

		vpc.subnets = subnets;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			vpc: vpc,
		});
	}

	onAddRoute = (i: number): void => {
		let vpc: VpcTypes.Vpc;

		vpc = {
			...this.state.vpc,
		};

		let routes = [
			...(vpc.routes || []),
		];

		routes.splice(i + 1, 0, {} as VpcTypes.Route);
		vpc.routes = routes;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			vpc: vpc,
		});
	}

	onChangeRoute(i: number, route: VpcTypes.Route): void {
		let vpc: VpcTypes.Vpc;

		vpc = {
			...this.state.vpc,
		};

		let routes = [
			...vpc.routes,
		];

		routes[i] = route;

		vpc.routes = routes;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			vpc: vpc,
		});
	}

	onRemoveRoute(i: number): void {
		let vpc: VpcTypes.Vpc;

		vpc = {
			...this.state.vpc,
		};

		let routes = [
			...vpc.routes,
		];

		routes.splice(i, 1);

		vpc.routes = routes;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			vpc: vpc,
		});
	}

	onAddMap = (i: number): void => {
		let vpc: VpcTypes.Vpc;

		vpc = {
			...this.state.vpc,
		};

		let maps = [
			...(vpc.maps || []),
		];

		if (maps.length === 0) {
			maps = [{}];
		}

		maps.splice(i + 1, 0, {} as VpcTypes.Map);
		vpc.maps = maps;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			vpc: vpc,
		});
	}

	onChangeMap(i: number, map: VpcTypes.Map): void {
		let vpc: VpcTypes.Vpc;

		vpc = {
			...this.state.vpc,
		};

		let maps = [
			...(vpc.maps || []),
		];

		if (maps.length === 0) {
			maps = [{}];
		}

		maps[i] = map;

		vpc.maps = maps;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			vpc: vpc,
		});
	}

	onRemoveMap(i: number): void {
		let vpc: VpcTypes.Vpc;

		vpc = {
			...this.state.vpc,
		};

		let maps = [
			...(vpc.maps || []),
		];

		if (maps.length !== 0) {
			maps.splice(i, 1);
		}

		if (maps.length === 0) {
			maps = [{}];
		}

		vpc.maps = maps;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			vpc: vpc,
		});
	}

	onAddArp = (i: number): void => {
		let vpc: VpcTypes.Vpc;

		vpc = {
			...this.state.vpc,
		};

		let arps = [
			...(vpc.arps || []),
		];

		if (arps.length === 0) {
			arps = [{}];
		}

		arps.splice(i + 1, 0, {} as VpcTypes.Arp);
		vpc.arps = arps;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			vpc: vpc,
		});
	}

	onChangeArp(i: number, arp: VpcTypes.Arp): void {
		let vpc: VpcTypes.Vpc;

		vpc = {
			...this.state.vpc,
		};

		let arps = [
			...(vpc.arps || []),
		];

		if (arps.length === 0) {
			arps = [{}];
		}

		arps[i] = arp;

		vpc.arps = arps;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			vpc: vpc,
		});
	}

	onRemoveArp(i: number): void {
		let vpc: VpcTypes.Vpc;

		vpc = {
			...this.state.vpc,
		};

		let arps = [
			...(vpc.arps || []),
		];

		if (arps.length !== 0) {
			arps.splice(i, 1);
		}

		if (arps.length === 0) {
			arps = [{}];
		}

		vpc.arps = arps;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			vpc: vpc,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let vpc: any = {
			...this.state.vpc,
		};

		if (this.props.organizations.length && !vpc.organization) {
			vpc.organization = this.props.organizations[0].id;
		}

		if (this.props.datacenters.length && !vpc.datacenter) {
			vpc.datacenter = this.props.datacenters[0].id;
		}

		VpcActions.create(vpc).then((): void => {
			this.setState({
				...this.state,
				message: 'VPC created successfully',
				changed: false,
			});

			setTimeout((): void => {
				this.setState({
					...this.state,
					disabled: false,
					changed: true,
				});
			}, 2000);
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let vpc: VpcTypes.Vpc = this.state.vpc;

		let hasOrganizations = false
		let organizationsSelect: JSX.Element[] = [];
		if (this.props.organizations.length) {
			for (let organization of this.props.organizations) {
				hasOrganizations = true
				organizationsSelect.push(
					<option
						key={organization.id}
						value={organization.id}
					>{organization.name}</option>,
				);
			}
		}

		if (!hasOrganizations) {
			organizationsSelect.push(
				<option key="null" value="">No Organizations</option>);
		}

		let hasDatacenters = false;
		let datacentersSelect: JSX.Element[] = [];
		if (this.props.datacenters.length) {
			datacentersSelect.push(
				<option key="null" value="">Select Datacenter</option>);

			hasDatacenters = true;
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

		let subnets = (vpc.subnets || []);
		if (subnets.length === 0) {
			subnets.push({});
		}

		let subnetsElem: JSX.Element[] = [];
		for (let i = 0; i < subnets.length; i++) {
			let index = i;

			subnetsElem.push(
				<VpcSubnet
					key={index}
					subnet={subnets[index]}
					onChange={(state: VpcTypes.Subnet): void => {
						this.onChangeSubnet(index, state);
					}}
					onAdd={(): void => {
						this.onAddSubnet(index);
					}}
					onRemove={(): void => {
						this.onRemoveSubnet(index);
					}}
				/>,
			);
		}

		let routes: JSX.Element[] = [
			<VpcRoute
				disabled={true}
				key={-1}
				route={{
					destination: '0.0.0.0/0',
					target: '0.0.0.0',
				} as VpcTypes.Route}
				onAdd={(): void => {
					this.onAddRoute(-1);
				}}
			/>,
		];
		if (vpc.routes) {
			for (let i = 0; i < (vpc.routes || []).length; i++) {
				let index = i;

				routes.push(
					<VpcRoute
						key={index}
						route={vpc.routes[index]}
						onChange={(state: VpcTypes.Route): void => {
							this.onChangeRoute(index, state);
						}}
						onAdd={(): void => {
							this.onAddRoute(index);
						}}
						onRemove={(): void => {
							this.onRemoveRoute(index);
						}}
					/>,
				);
			}
		}

		let maps = (vpc.maps || []);
		if (maps.length === 0) {
			maps.push({});
		}

		let mapsElem: JSX.Element[] = [];
		for (let i = 0; i < maps.length; i++) {
			let index = i;

			mapsElem.push(
				<VpcMap
					key={index}
					map={maps[index]}
					onChange={(state: VpcTypes.Map): void => {
						this.onChangeMap(index, state);
					}}
					onAdd={(): void => {
						this.onAddMap(index);
					}}
					onRemove={(): void => {
						this.onRemoveMap(index);
					}}
				/>,
			);
		}

		let arps = (vpc.arps || []);
		if (arps.length === 0) {
			arps.push({});
		}

		let arpsElem: JSX.Element[] = [];
		for (let i = 0; i < arps.length; i++) {
			let index = i;

			arpsElem.push(
				<VpcArp
					key={index}
					arp={arps[index]}
					onChange={(state: VpcTypes.Arp): void => {
						this.onChangeArp(index, state);
					}}
					onAdd={(): void => {
						this.onAddArp(index);
					}}
					onRemove={(): void => {
						this.onRemoveArp(index);
					}}
				/>,
			);
		}

		return <div
			className="bp5-card bp5-row"
			style={css.row}
		>
			<td
				className="bp5-cell"
				colSpan={4}
				style={css.card}
			>
				<div className="layout horizontal wrap">
					<div style={css.group}>
						<div style={css.buttons}>
						</div>
						<PageInput
							label="Name"
							help="Name of vpc"
							type="text"
							placeholder="Enter name"
							value={vpc.name}
							onChange={(val): void => {
								this.set('name', val);
							}}
						/>
						<PageTextArea
							label="Comment"
							help="VPC comment."
							placeholder="VPC comment"
							rows={3}
							value={vpc.comment}
							onChange={(val: string): void => {
								this.set('comment', val);
							}}
						/>
						<PageInput
							label="Network"
							help="Network address of vpc with cidr."
							type="text"
							placeholder="Enter network"
							value={vpc.network}
							onChange={(val): void => {
								this.set('network', val);
							}}
						/>
						<label style={css.itemsLabel}>
							Subnets
							<Help
								title="Subnets"
								content="Subnets in VPC, can only be added or removed. Once added a subnet network block cannot be modified."
							/>
						</label>
						<div style={css.list}>
							{subnetsElem}
						</div>
						<label style={css.itemsLabel}>
							Network Maps
							<Help
								title="Network Maps"
								content="Map destination network CIDR to new target IP."
							/>
						</label>
						<div style={css.list}>
							{mapsElem}
						</div>
					</div>
					<div style={css.group}>
						<PageSelect
							disabled={this.state.disabled}
							hidden={Constants.user}
							label="Organization"
							help="Organization for VPC."
							value={vpc.organization}
							onChange={(val): void => {
								this.set('organization', val);
							}}
						>
							{organizationsSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasDatacenters}
							label="Datacenter"
							help="Datacenter for VPC."
							value={vpc.datacenter}
							onChange={(val): void => {
								this.set('datacenter', val);
							}}
						>
							{datacentersSelect}
						</PageSelect>
						<PageSwitch
							disabled={this.state.disabled}
							label="ICMP Redirects"
							help="Enable or disable ICMP redirects for VPC routing table. ICMP redirects will improve the routing path of static routes in the VPC routing table but will be cached by the instance for 5 minutes unless adjusted on the system. If dynamic updates to the VPC routing table are made such as with failover site-to-site systems redirects should be disabled to allow fast failover to the new route. ICMP redirects are not recommended for most configurations."
							checked={vpc.icmp_redirects}
							onToggle={(): void => {
								this.set('icmp_redirects', !vpc.icmp_redirects);
							}}
						/>
						<label style={css.itemsLabel}>
							Route Table
							<Help
								title="Route Table"
								content="VPC routing table, enter a CIDR network for the desitnation and IP address for taget."
							/>
						</label>
						<div style={css.list}>
							{routes}
						</div>
						<label style={css.itemsLabel}>
							Custom ARP
							<Help
								title="Custom ARP"
								content="Custom ARP entries for external resources on VPC VLAN."
							/>
						</label>
						<div style={css.list}>
							{arpsElem}
						</div>
					</div>
				</div>
				<PageCreate
					style={css.save}
					hidden={!this.state.vpc}
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
