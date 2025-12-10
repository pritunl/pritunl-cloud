/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as VpcTypes from '../types/VpcTypes';
import * as VpcActions from '../actions/VpcActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import * as PageInfos from './PageInfo';
import CompletionStore from "../stores/CompletionStore";
import VpcRoute from './VpcRoute';
import VpcMap from './VpcMap';
import VpcArp from './VpcArp';
import VpcSubnet from './VpcSubnet';
import PageInput from './PageInput';
import PageSwitch from './PageSwitch';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import PageTextArea from "./PageTextArea";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	vpc: VpcTypes.VpcRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	addNetworkRole: string;
	addVpc: string;
	vpc: VpcTypes.Vpc;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 0 10px',
		width: '100%',
	} as React.CSSProperties,
	button: {
		height: '30px',
	} as React.CSSProperties,
	buttons: {
		cursor: 'pointer',
		position: 'absolute',
		top: 0,
		left: 0,
		right: 0,
		padding: '4px',
		height: '39px',
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
		height: '20px',
	} as React.CSSProperties,
	list: {
		marginBottom: '15px',
	} as React.CSSProperties,
};

export default class VpcDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			addNetworkRole: null,
			addVpc: null,
			vpc: null,
		};
	}

	set(name: string, val: any): void {
		let vpc: any;

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

		vpc[name] = val;

		this.setState({
			...this.state,
			changed: true,
			vpc: vpc,
		});
	}

	onAddSubnet = (i: number): void => {
		let vpc: VpcTypes.Vpc;

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

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

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

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

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

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

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

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

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

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

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

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

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

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

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

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

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

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

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

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

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

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

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

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

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		VpcActions.commit(this.state.vpc).then((): void => {
			this.setState({
				...this.state,
				message: 'Your changes have been saved',
				changed: false,
				disabled: false,
			});

			setTimeout((): void => {
				if (!this.state.changed) {
					this.setState({
						...this.state,
						vpc: null,
						changed: false,
					});
				}
			}, 1000);

			setTimeout((): void => {
				if (!this.state.changed) {
					this.setState({
						...this.state,
						message: '',
					});
				}
			}, 3000);
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		VpcActions.remove(this.props.vpc.id).then((): void => {
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
	}

	render(): JSX.Element {
		let vpc: VpcTypes.Vpc = this.state.vpc ||
			this.props.vpc;

		let datacenter = CompletionStore.datacenter(vpc.datacenter);
		let datacenterName = datacenter ? datacenter.name : vpc.datacenter;
		let org = CompletionStore.organization(this.props.vpc.organization);

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

		let fields: PageInfos.Field[] = [
			{
				label: 'ID',
				value: this.props.vpc.id || 'Unknown',
			},
			{
				label: 'Datacenter',
				value: datacenterName,
			},
			{
				label: 'Organization',
				value: org ? org.name : this.props.vpc.organization,
			},
			{
				label: 'VLAN Number',
				value: this.props.vpc.vpc_id || 'Unknown',
			},
			{
				label: 'Private IPv6 Network',
				value: this.props.vpc.network6 || 'Unknown',
				copy: true,
			},
		];

		return <td
			className="bp5-cell"
			colSpan={4}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal tab-close"
						style={css.buttons}
						onClick={(evt): void => {
							let target = evt.target as HTMLElement;

							if (target.className.indexOf('tab-close') !== -1) {
								this.props.onClose();
							}
						}}
					>
						<div>
							<label
								className="bp5-control bp5-checkbox"
								style={css.select}
							>
								<input
									type="checkbox"
									checked={this.props.selected}
									onChange={(evt): void => {
									}}
									onClick={(evt): void => {
										this.props.onSelect(evt.shiftKey);
									}}
								/>
								<span className="bp5-control-indicator"/>
							</label>
						</div>
						<div className="flex tab-close"/>
						<ConfirmButton
							className="bp5-minimal bp5-intent-danger bp5-icon-trash"
							style={css.button}
							safe={true}
							progressClassName="bp5-intent-danger"
							dialogClassName="bp5-intent-danger bp5-icon-delete"
							dialogLabel="Delete VPC"
							confirmMsg="Permanently delete this VPC"
							confirmInput={true}
							items={[vpc.name]}
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
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
					<PageInfo
						fields={fields}
					/>
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
			<PageSave
				style={css.save}
				hidden={!this.state.vpc && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						vpc: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}
