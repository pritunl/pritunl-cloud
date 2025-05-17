/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as NodeTypes from '../types/NodeTypes';
import * as CertificateTypes from '../types/CertificateTypes';
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as ZoneTypes from '../types/ZoneTypes';
import * as NodeActions from '../actions/NodeActions';
import * as BlockTypes from '../types/BlockTypes';
import * as MiscUtils from '../utils/MiscUtils';
import * as PageInfos from './PageInfo';
import CertificatesStore from '../stores/CertificatesStore';
import NodeDeploy from './NodeDeploy';
import PageInput from './PageInput';
import PageSwitch from './PageSwitch';
import PageInputSwitch from './PageInputSwitch';
import PageSelect from './PageSelect';
import PageSelectButton from './PageSelectButton';
import PageInputButton from './PageInputButton';
import PageTextArea from './PageTextArea';
import PageNumInput from './PageNumInput';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import NodeBlock from './NodeBlock';
import ConfirmButton from './ConfirmButton';
import Help from './Help';

interface Props {
	node: NodeTypes.NodeRo;
	certificates: CertificateTypes.CertificatesRo;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
	blocks: BlockTypes.BlocksRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	datacenter: string;
	zone: string;
	changed: boolean;
	message: string;
	node: NodeTypes.Node;
	addExternalIface: string;
	addInternalIface: string;
	addOracleSubnet: string;
	addCert: string;
	addNetworkRole: string;
	addDrive: string;
	forwardedChecked: boolean;
	forwardedProtoChecked: boolean;
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
		backgroundColor: 'rgba(0, 0, 0, 0.13)',
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		wordBreak: 'break-all',
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
	restart: {
		marginRight: '10px',
	} as React.CSSProperties,
	label: {
		width: '100%',
		maxWidth: '280px',
	} as React.CSSProperties,
	labelWide: {
		width: '100%',
		maxWidth: '400px',
	} as React.CSSProperties,
	inputGroup: {
		width: '100%',
	} as React.CSSProperties,
	protocol: {
		minWidth: '90px',
		flex: '0 1 auto',
	} as React.CSSProperties,
	port: {
		minWidth: '120px',
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
	blocks: {
		marginBottom: '15px',
	} as React.CSSProperties,
};

export default class NodeDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			datacenter: '',
			zone: '',
			changed: false,
			message: '',
			node: null,
			addExternalIface: null,
			addInternalIface: null,
			addOracleSubnet: null,
			addCert: null,
			addNetworkRole: null,
			addDrive: null,
			forwardedChecked: false,
			forwardedProtoChecked: false,
		};
	}

	set(name: string, val: any): void {
		let node: any;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		node[name] = val;

		this.setState({
			...this.state,
			changed: true,
			node: node,
		});
	}

	toggleFirewall(): void {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		node.firewall = !node.firewall;
		if (!node.firewall) {
			node.network_roles = [];
		}

		this.setState({
			...this.state,
			changed: true,
			node: node,
		});
	}

	toggleType(typ: string): void {
		let node: NodeTypes.Node = this.state.node || this.props.node;

		let vals = node.types;

		let i = vals.indexOf(typ);
		if (i === -1) {
			vals.push(typ);
		} else {
			vals.splice(i, 1);
		}

		vals = vals.filter((val): boolean => {
			return !!val;
		});

		vals.sort();

		this.set('types', vals);
	}

	ifaces(): string[] {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let dcId = node.datacenter;
		if (this.state.datacenter) {
			dcId = this.state.datacenter;
		}

		let vxlan = false;
		for (let dc of this.props.datacenters) {
			if (dc.id === dcId) {
				if (dc.network_mode === 'vxlan_vlan') {
					vxlan = true;
				}
				break;
			}
		}

		if (vxlan) {
			return node.available_bridges.concat(node.available_interfaces);
		} else {
			return node.available_bridges.concat(node.available_interfaces);
		}
	}

	subnetLabel(subnetId: string): string {
		for (let vpc of (this.props.node.available_vpcs || [])) {
			for (let subnet of (vpc.subnets || [])) {
				if (subnet.id === subnetId) {
					return vpc.name + ' - ' + subnet.name;
				}
			}
		}
		return subnetId;
	}

	onAddNetworkRole = (): void => {
		let node: NodeTypes.Node;

		if (!this.state.addNetworkRole) {
			return;
		}

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let networkRoles = [
			...(node.network_roles || []),
		];

		if (networkRoles.indexOf(this.state.addNetworkRole) === -1) {
			networkRoles.push(this.state.addNetworkRole);
		}

		networkRoles.sort();
		node.network_roles = networkRoles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addNetworkRole: '',
			node: node,
		});
	}

	onRemoveNetworkRole = (networkRole: string): void => {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let networkRoles = [
			...(node.network_roles || []),
		];

		let i = networkRoles.indexOf(networkRole);
		if (i === -1) {
			return;
		}

		networkRoles.splice(i, 1);
		node.network_roles = networkRoles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addNetworkRole: '',
			node: node,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let node = {
			...this.state.node,
		};

		if (!this.props.node.zone) {
			let zone = this.state.zone;
			if (!zone && this.props.datacenters.length &&
					this.props.zones.length) {
				let datacenter = this.state.datacenter ||
					this.props.datacenters[0].id;
				for (let zne of this.props.zones) {
					if (zne.datacenter === datacenter) {
						zone = zne.id;
					}
				}
			}

			if (zone) {
				node.zone = zone;
			}
		}

		NodeActions.commit(node).then((): void => {
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
						message: '',
						changed: false,
						node: null,
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

	operation(state: string): void {
		this.setState({
			...this.state,
			disabled: true,
		});
		NodeActions.operation(this.props.node.id, state).then((): void => {
			setTimeout((): void => {
				this.setState({
					...this.state,
					disabled: false,
				});
			}, 250);
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		NodeActions.remove(this.props.node.id).then((): void => {
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

	onAddExternalIface = (): void => {
		let node: NodeTypes.Node;
		let availableIfaces = this.ifaces();

		if (!this.state.addExternalIface && !availableIfaces.length) {
			return;
		}

		let index = this.state.addExternalIface || availableIfaces[0];

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let ifaces = [
			...(node.external_interfaces || []),
		];

		if (ifaces.indexOf(index) === -1) {
			ifaces.push(index);
		}

		ifaces.sort();

		node.external_interfaces = ifaces;

		this.setState({
			...this.state,
			changed: true,
			node: node,
		});
	}

	onRemoveExternalIface = (iface: string): void => {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let ifaces = [
			...(node.external_interfaces || []),
		];

		let i = ifaces.indexOf(iface);
		if (i === -1) {
			return;
		}

		ifaces.splice(i, 1);

		node.external_interfaces = ifaces;

		this.setState({
			...this.state,
			changed: true,
			node: node,
		});
	}

	onAddInternalIface = (): void => {
		let node: NodeTypes.Node;
		let availableIfaces = this.ifaces();

		if (!this.state.addInternalIface && !availableIfaces.length) {
			return;
		}

		let index = this.state.addInternalIface || availableIfaces[0];

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let ifaces = [
			...(node.internal_interfaces || []),
		];

		if (ifaces.indexOf(index) === -1) {
			ifaces.push(index);
		}

		ifaces.sort();

		node.internal_interfaces = ifaces;

		this.setState({
			...this.state,
			changed: true,
			node: node,
		});
	}

	onRemoveInternalIface = (iface: string): void => {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let ifaces = [
			...(node.internal_interfaces || []),
		];

		let i = ifaces.indexOf(iface);
		if (i === -1) {
			return;
		}

		ifaces.splice(i, 1);

		node.internal_interfaces = ifaces;

		this.setState({
			...this.state,
			changed: true,
			node: node,
		});
	}

	onAddCert = (): void => {
		let node: NodeTypes.Node;

		if (!this.state.addCert && !this.props.certificates.length) {
			return;
		}

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let certId = this.state.addCert;
		if (!certId) {
			for (let certificate of this.props.certificates) {
				if (certificate.organization) {
					continue;
				}
				certId = certificate.id;
				break;
			}
		}

		let certificates = [
			...(node.certificates || []),
		];

		if (certificates.indexOf(certId) === -1) {
			certificates.push(certId);
		}

		certificates.sort();

		node.certificates = certificates;

		this.setState({
			...this.state,
			changed: true,
			node: node,
		});
	}

	onRemoveCert = (certId: string): void => {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let certificates = [
			...(node.certificates || []),
		];

		let i = certificates.indexOf(certId);
		if (i === -1) {
			return;
		}

		certificates.splice(i, 1);

		node.certificates = certificates;

		this.setState({
			...this.state,
			changed: true,
			node: node,
		});
	}

	newBlock = (ipv6: boolean): NodeTypes.BlockAttachment => {
		let defBlock = '';

		for (let block of (this.props.blocks || [])) {
			if ((ipv6 && block.type === 'ipv6') ||
					(!ipv6 && block.type === 'ipv4')) {
				defBlock = block.id;
			}
		}

		return {
			interface: this.props.node.available_bridges.concat(
				this.props.node.available_interfaces)[0],
			block: defBlock,
		} as NodeTypes.BlockAttachment;
	}

	onNetworkMode = (mode: string): void => {
		let node: any;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		if (mode === 'static' && (node.blocks || []).length === 0) {
			node.blocks = [
				this.newBlock(false),
			];
		}

		node.network_mode = mode;

		this.setState({
			...this.state,
			changed: true,
			node: node,
		});
	}

	onNetworkMode6 = (mode: string): void => {
		let node: any;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		if (mode === 'static' && (node.blocks6 || []).length === 0) {
			node.blocks6 = [
				this.newBlock(true),
			];
		}

		node.network_mode6 = mode;

		this.setState({
			...this.state,
			changed: true,
			node: node,
		});
	}

	onAddOracleSubnet = (): void => {
		let node: NodeTypes.Node;
		let availabeVpcs = this.props.node.available_vpcs || [];

		if (!this.state.addOracleSubnet && !availabeVpcs.length &&
				!availabeVpcs[0].subnets.length) {
			return;
		}

		let addOracleSubnet = this.state.addOracleSubnet;
		if (!addOracleSubnet) {
			addOracleSubnet = availabeVpcs[0].subnets[0].id;
		}

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let nodeOracleSubnets = [
			...(node.oracle_subnets || []),
		];

		let index = -1;
		for (let i = 0; i < nodeOracleSubnets.length; i++) {
			if (nodeOracleSubnets[i] === addOracleSubnet) {
				index = i;
				break
			}
		}

		if (index === -1) {
			nodeOracleSubnets.push(addOracleSubnet);
		}

		node.oracle_subnets = nodeOracleSubnets;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			node: node,
		});
	}

	onRemoveOracleSubnet = (device: string): void => {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let nodeOracleSubnets = [
			...(node.oracle_subnets || []),
		];

		let index = -1;
		for (let i = 0; i < nodeOracleSubnets.length; i++) {
			if (nodeOracleSubnets[i] === device) {
				index = i;
				break
			}
		}
		if (index === -1) {
			return;
		}

		nodeOracleSubnets.splice(index, 1);
		node.oracle_subnets = nodeOracleSubnets;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			node: node,
		});
	}

	onAddBlock = (i: number): void => {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let blocks = [
			...node.blocks,
		];

		blocks.splice(i + 1, 0, this.newBlock(false));
		node.blocks = blocks;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			node: node,
		});
	}

	onChangeBlock(i: number, block: NodeTypes.BlockAttachment): void {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let blocks = [
			...node.blocks,
		];

		blocks[i] = block;

		node.blocks = blocks;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			node: node,
		});
	}

	onRemoveBlock(i: number): void {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let blocks = [
			...node.blocks,
		];

		blocks.splice(i, 1);

		if (!blocks.length) {
			blocks = [
				this.newBlock(false),
			];
		}

		node.blocks = blocks;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			node: node,
		});
	}

	onAddBlock6 = (i: number): void => {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let blocks = [
			...node.blocks6,
		];

		blocks.splice(i + 1, 0, this.newBlock(true));
		node.blocks6 = blocks;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			node: node,
		});
	}

	onChangeBlock6(i: number, block: NodeTypes.BlockAttachment): void {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let blocks = [
			...node.blocks6,
		];

		blocks[i] = block;

		node.blocks6 = blocks;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			node: node,
		});
	}

	onRemoveBlock6(i: number): void {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let blocks = [
			...node.blocks6,
		];

		blocks.splice(i, 1);

		if (!blocks.length) {
			blocks = [
				this.newBlock(true),
			];
		}

		node.blocks6 = blocks;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			node: node,
		});
	}

	onAddDrive = (): void => {
		let node: NodeTypes.Node;
		let availabeDrives = this.props.node.available_drives || [];

		if (!this.state.addDrive && !availabeDrives.length) {
			return;
		}

		let addDrive = this.state.addDrive;
		if (!addDrive) {
			addDrive = availabeDrives[0].id;
		}

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let instanceDrives = [
			...(node.instance_drives || []),
		];

		let index = -1;
		for (let i = 0; i < instanceDrives.length; i++) {
			let dev = instanceDrives[i];
			if (dev.id === addDrive) {
				index = i;
				break
			}
		}

		if (index === -1) {
			instanceDrives.push({
				id: addDrive,
			});
		}

		node.instance_drives = instanceDrives;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			node: node,
		});
	}

	onRemoveDrive = (device: string): void => {
		let node: NodeTypes.Node;

		if (this.state.changed) {
			node = {
				...this.state.node,
			};
		} else {
			node = {
				...this.props.node,
			};
		}

		let instanceDrives = [
			...(node.instance_drives || []),
		];

		let index = -1;
		for (let i = 0; i < instanceDrives.length; i++) {
			let dev = instanceDrives[i];
			if (dev.id === device) {
				index = i;
				break
			}
		}
		if (index === -1) {
			return;
		}

		instanceDrives.splice(index, 1);
		node.instance_drives = instanceDrives;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			node: node,
		});
	}

	render(): JSX.Element {
		let node: NodeTypes.Node = this.state.node || this.props.node;
		let active = node.requests_min !== 0 || node.memory !== 0 ||
				node.load1 !== 0 || node.load5 !== 0 || node.load15 !== 0;
		let types = node.types || [];

		let publicIps: any = this.props.node.public_ips;
		if (!publicIps || !publicIps.length) {
			publicIps = 'None';
		}

		let publicIps6: any = this.props.node.public_ips6;
		if (!publicIps6 || !publicIps6.length) {
			publicIps6 = 'None';
		}

		let privateIps: any = this.props.node.private_ips;
		if (!privateIps || !privateIps.length) {
			privateIps = 'None';
		}

		let resourceBars: PageInfos.Bar[] = [
			{
				progressClass: 'bp5-no-stripes bp5-intent-success',
				label: 'Load1',
				value: this.props.node.load1 || 0,
			},
			{
				progressClass: 'bp5-no-stripes bp5-intent-warning',
				label: 'Load5',
				value: this.props.node.load5 || 0,
			},
			{
				progressClass: 'bp5-no-stripes bp5-intent-danger',
				label: 'Load15',
				value: this.props.node.load15 || 0,
			},
			{
				progressClass: 'bp5-no-stripes bp5-intent-primary',
				label: 'Memory',
				value: this.props.node.memory || 0,
			},
		];
		if (this.props.node.hugepages) {
			resourceBars.push({
				progressClass: 'bp5-no-stripes bp5-intent-primary',
				label: 'HugePages',
				value: this.props.node.hugepages_used || 0,
				color: '#7207d4',
			});
		}

		let externalIfaces: JSX.Element[] = [];
		for (let iface of (node.external_interfaces || [])) {
			externalIfaces.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.item}
					key={iface}
				>
					{iface}
					<button
						disabled={this.state.disabled}
						className="bp5-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveExternalIface(iface);
						}}
					/>
				</div>,
			);
		}

		let internalIfaces: JSX.Element[] = [];
		for (let iface of (node.internal_interfaces || [])) {
			internalIfaces.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.item}
					key={iface}
				>
					{iface}
					<button
						disabled={this.state.disabled}
						className="bp5-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveInternalIface(iface);
						}}
					/>
				</div>,
			);
		}

		let availableIfaces = this.ifaces();
		let externalIfacesSelect: JSX.Element[] = [];
		for (let iface of (availableIfaces || [])) {
			externalIfacesSelect.push(
				<option key={iface} value={iface}>
					{iface}
				</option>,
			);
		}

		let internalIfacesSelect: JSX.Element[] = [];
		for (let iface of (availableIfaces || [])) {
			internalIfacesSelect.push(
				<option key={iface} value={iface}>
					{iface}
				</option>,
			);
		}

		let oracleSubnets: JSX.Element[] = [];
		for (let subnetId of (node.oracle_subnets || [])) {
			oracleSubnets.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.item}
					key={subnetId}
				>
					{this.subnetLabel(subnetId)}
					<button
						disabled={this.state.disabled}
						className="bp5-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveOracleSubnet(subnetId);
						}}
					/>
				</div>,
			);
		}

		let availableSubnetsSelect: JSX.Element[] = [];
		for (let vpc of (node.available_vpcs || [])) {
			for (let subnet of (vpc.subnets || [])) {
				availableSubnetsSelect.push(
					<option key={subnet.id} value={subnet.id}>
						{vpc.name + ' - ' + subnet.name}
					</option>,
				);
			}
		}

		let availableDrives: JSX.Element[] = [];
		for (let device of (node.instance_drives || [])) {
			availableDrives.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.item}
					key={device.id}
				>
					{device.id}
					<button
						disabled={this.state.disabled}
						className="bp5-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveDrive(device.id);
						}}
					/>
				</div>,
			);
		}

		let availableDrivesSelect: JSX.Element[] = [];
		for (let device of (node.available_drives || [])) {
			availableDrivesSelect.push(
				<option key={device.id} value={device.id}>
					{device.id}
				</option>,
			);
		}

		let certificates: JSX.Element[] = [];
		for (let certId of (node.certificates || [])) {
			let cert = CertificatesStore.certificate(certId);
			if (!cert) {
				continue;
			}

			certificates.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.item}
					key={cert.id}
				>
					{cert.name}
					<button
						disabled={this.state.disabled}
						className="bp5-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveCert(cert.id);
						}}
					/>
				</div>,
			);
		}

		let hasCertificates = false;
		let certificatesSelect: JSX.Element[] = [];
		if (this.props.certificates.length) {
			for (let certificate of this.props.certificates) {
				if (certificate.organization) {
					continue;
				}
				hasCertificates = true;

				certificatesSelect.push(
					<option key={certificate.id} value={certificate.id}>
						{certificate.name}
					</option>,
				);
			}
		}

		if (!hasCertificates) {
			certificatesSelect = [
				<option key="null" value="">
					No Certificates
				</option>,
			];
		}

		let defaultDatacenter = '';
		let hasDatacenters = false;
		let datacentersSelect: JSX.Element[] = [];
		if (this.props.datacenters.length) {
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
				if (!this.props.node.zone && zone.datacenter !== datacenter) {
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

		let hasRenders = false;
		let rendersSelect: JSX.Element[] = [];
		if (this.props.node.available_renders &&
			this.props.node.available_renders.length) {
			rendersSelect.push(<option key="null" value="">Select Render</option>);

			for (let render of this.props.node.available_renders) {
				hasRenders = true;

				rendersSelect.push(
					<option
						key={render}
						value={render}
					>{render}</option>,
				);
			}
		}

		if (!hasRenders) {
			rendersSelect = [<option key="null" value="">No Renders</option>];
		}

		let networkRoles: JSX.Element[] = [];
		for (let networkRole of (node.network_roles || [])) {
			networkRoles.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.role}
					key={networkRole}
				>
					{networkRole}
					<button
						className="bp5-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveNetworkRole(networkRole);
						}}
					/>
				</div>,
			);
		}

		let nodeBlocks = node.blocks || [];
		let blocks: JSX.Element[] = [];
		for (let i = 0; i < nodeBlocks.length; i++) {
			let index = i;

			blocks.push(
				<NodeBlock
					key={index}
					interfaces={node.available_bridges.concat(
						node.available_interfaces)}
					blocks={this.props.blocks}
					block={nodeBlocks[index]}
					ipv6={false}
					onChange={(state: NodeTypes.BlockAttachment): void => {
						this.onChangeBlock(index, state);
					}}
					onAdd={(): void => {
						this.onAddBlock(index);
					}}
					onRemove={(): void => {
						this.onRemoveBlock(index);
					}}
				/>,
			);
		}

		let nodeBlocks6 = node.blocks6 || [];
		let blocks6: JSX.Element[] = [];
		for (let i = 0; i < nodeBlocks6.length; i++) {
			let index = i;

			blocks6.push(
				<NodeBlock
					key={index}
					interfaces={node.available_bridges.concat(
						node.available_interfaces)}
					blocks={this.props.blocks}
					block={nodeBlocks6[index]}
					ipv6={true}
					onChange={(state: NodeTypes.BlockAttachment): void => {
						this.onChangeBlock6(index, state);
					}}
					onAdd={(): void => {
						this.onAddBlock6(index);
					}}
					onRemove={(): void => {
						this.onRemoveBlock6(index);
					}}
				/>,
			);
		}

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
						<div className="flex tab-close"/>
						<ConfirmButton
							className="bp5-minimal bp5-intent-danger bp5-icon-trash"
							style={css.button}
							safe={true}
							progressClassName="bp5-intent-danger"
							dialogClassName="bp5-intent-danger bp5-icon-delete"
							dialogLabel="Delete Node"
							confirmMsg="Permanently delete this node"
							confirmInput={true}
							items={[node.name]}
							disabled={active || this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						disabled={this.state.disabled}
						label="Name"
						help="Name of node"
						type="text"
						placeholder="Enter name"
						value={node.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageTextArea
						label="Comment"
						help="Node comment."
						placeholder="Node comment"
						rows={3}
						value={node.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Admin"
						help="Provides access to the admin console on this node."
						checked={types.indexOf('admin') !== -1}
						onToggle={(): void => {
							this.toggleType('admin');
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="User"
						help="Provides access to the user console on this node for SSH certificates."
						checked={types.indexOf('user') !== -1}
						onToggle={(): void => {
							this.toggleType('user');
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Load Balancer"
						help="Provides access to load balancers."
						checked={types.indexOf('balancer') !== -1}
						onToggle={(): void => {
							this.toggleType('balancer');
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Hypervisor"
						help="Run instances with hypervisor on this node."
						checked={types.indexOf('hypervisor') !== -1}
						onToggle={(): void => {
							this.toggleType('hypervisor');
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={types.indexOf('balancer') === -1 && (
							types.indexOf('admin') === -1 ||
							types.indexOf('user') === -1)}
						label="Admin Domain"
						help="Domain that will be used to access the admin interface."
						type="text"
						placeholder="Enter admin domain"
						value={node.admin_domain}
						onChange={(val): void => {
							this.set('admin_domain', val);
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={types.indexOf('balancer') === -1 && (
							types.indexOf('admin') === -1 ||
							types.indexOf('user') === -1)}
						label="User Domain"
						help="Domain that will be used to access the user interface."
						type="text"
						placeholder="Enter user domain"
						value={node.user_domain}
						onChange={(val): void => {
							this.set('user_domain', val);
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={types.indexOf('admin') === -1 &&
							types.indexOf('user') === -1}
						label="WebAuthn Domain"
						help="Domain that will be used for WebAuthn relying party identifier. This domain should be the highest level domain for the relevant resources. All other Pritunl Cloud domains handling WebAuthn authentication must be a sub-domain of this domain. Changing this domain will invalidate all existing WebAuthn devices."
						type="text"
						placeholder="Enter WebAuthn domain"
						value={node.webauthn_domain}
						onChange={(val): void => {
							this.set('webauthn_domain', val);
						}}
					/>
					<label className="bp5-label" style={css.label}>
						Protocol and Port
						<div className="bp5-control-group" style={css.inputGroup}>
							<div className="bp5-select" style={css.protocol}>
								<select
									disabled={this.state.disabled}
									value={node.protocol || 'https'}
									onChange={(evt): void => {
										this.set('protocol', evt.target.value);
									}}
								>
									<option value="http">HTTP</option>
									<option value="https">HTTPS</option>
								</select>
							</div>
							<input
								className="bp5-input"
								disabled={this.state.disabled}
								style={css.port}
								type="text"
								autoCapitalize="off"
								spellCheck={false}
								placeholder="Port"
								value={node.port || 443}
								onChange={(evt): void => {
									this.set('port', parseInt(evt.target.value, 10));
								}}
							/>
						</div>
					</label>
					<PageSwitch
						disabled={this.state.disabled}
						label="Web redirect server"
						help="Enable redirect server for HTTP requests to HTTPS. Required for Lets Encrypt certificates."
						checked={!node.no_redirect_server}
						onToggle={(): void => {
							this.set('no_redirect_server', !node.no_redirect_server);
						}}
					/>
					<PageSelect
						disabled={this.state.disabled || !hasDatacenters}
						hidden={!!this.props.node.zone}
						label="Datacenter"
						help="Node datacenter, cannot be changed once set."
						value={this.state.datacenter}
						onChange={(val): void => {
							if (this.state.changed) {
								node = {
									...this.state.node,
								};
							} else {
								node = {
									...this.props.node,
								};
							}

							this.setState({
								...this.state,
								changed: true,
								node: node,
								datacenter: val,
								zone: '',
							});
						}}
					>
						{datacentersSelect}
					</PageSelect>
					<PageSelect
						disabled={!!this.props.node.zone || this.state.disabled ||
							!hasZones}
						label="Zone"
						help="Node zone, cannot be changed once set. Clear node ID in configuration file to reset node."
						value={this.props.node.zone ? this.props.node.zone :
							this.state.zone}
						onChange={(val): void => {
							let node: NodeTypes.Node;
							if (this.state.changed) {
								node = {
									...this.state.node,
								};
							} else {
								node = {
									...this.props.node,
								};
							}

							this.setState({
								...this.state,
								changed: true,
								node: node,
								zone: val,
							});
						}}
					>
						{zonesSelect}
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled}
						label="Network IPv4 Mode"
						help="Network mode for public IP addresses. Cannot be changed with instances running."
						value={node.network_mode}
						onChange={(val): void => {
							this.onNetworkMode(val);
						}}
					>
						<option value="dhcp">DHCP</option>
						<option value="static">Static</option>
						<option value="oracle">Oracle Cloud</option>
						<option value="disabled">Disabled</option>
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled}
						label="Network IPv6 Mode"
						help="Network mode for public IPv6 addresses. Cannot be changed with instances running. Default will use IPv4 network mode."
						value={node.network_mode6}
						onChange={(val): void => {
							this.onNetworkMode6(val);
						}}
					>
						<option value="dhcp">DHCP</option>
						<option value="static">Static</option>
						<option value="oracle">Oracle Cloud</option>
						<option value="disabled">Disabled</option>
					</PageSelect>
					<label
						className="bp5-label"
						style={css.label}
						hidden={
							node.network_mode !== 'dhcp' &&
							node.network_mode !== '' &&
							node.network_mode6 !== 'dhcp' &&
							node.network_mode6 !== ''
						}
					>
						External Interfaces
						<Help
							title="External Interfaces"
							content="External interfaces for instance public interface, must be a bridge interface. Leave blank for automatic configuration."
						/>
						<div>
							{externalIfaces}
						</div>
					</label>
					<PageSelectButton
						hidden={
							node.network_mode !== 'dhcp' &&
							node.network_mode !== '' &&
							node.network_mode6 !== 'dhcp' &&
							node.network_mode6 !== ''
						}
						label="Add Interface"
						value={this.state.addExternalIface}
						disabled={!externalIfacesSelect.length || this.state.disabled}
						buttonClass="bp5-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								addExternalIface: val,
							});
						}}
						onSubmit={this.onAddExternalIface}
					>
						{externalIfacesSelect}
					</PageSelectButton>
					<label
						className="bp5-label"
						style={css.label}
					>
						Internal Interfaces
						<Help
							title="Internal Interfaces"
							content="Internal interfaces for instance private VPC interface. If zone network mode is default this must be a bridge interface. Set zone network mode to VXLan to use non-bridge interface. Leave blank to use external interface."
						/>
						<div>
							{internalIfaces}
						</div>
					</label>
					<PageSelectButton
						label="Add Interface"
						value={this.state.addInternalIface}
						disabled={!internalIfacesSelect.length || this.state.disabled}
						buttonClass="bp5-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								addInternalIface: val,
							});
						}}
						onSubmit={this.onAddInternalIface}
					>
						{internalIfacesSelect}
					</PageSelectButton>
					<label
						className="bp5-label"
						hidden={node.network_mode !== 'static'}
						style={css.labelWide}
					>
						External IPv4 Block Attachments
						{blocks}
					</label>
					<label
						className="bp5-label"
						hidden={node.network_mode6 !== 'static'}
						style={css.labelWide}
					>
						External IPv6 Block Attachments
						{blocks6}
					</label>
					<label
						className="bp5-label"
						hidden={node.network_mode !== 'oracle'}
						style={css.label}
					>
						Oracle Cloud Subnets
						<Help
							title="Oracle Cloud Subnets"
							content="Oracle Cloud VCN subnets available to attach to instances."
						/>
						<div>
							{oracleSubnets}
						</div>
					</label>
					<PageSelectButton
						label="Add Subnet"
						hidden={node.network_mode !== 'oracle'}
						value={this.state.addOracleSubnet}
						disabled={!availableSubnetsSelect.length || this.state.disabled}
						buttonClass="bp5-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								addOracleSubnet: val,
							});
						}}
						onSubmit={this.onAddOracleSubnet}
					>
						{availableSubnetsSelect}
					</PageSelectButton>
					<PageSwitch
						disabled={this.state.disabled}
						label="Host Network"
						help="Enable host networking to allow host to instance communication. Required for instance NAT."
						checked={!node.no_host_network}
						onToggle={(): void => {
							this.set('no_host_network', !node.no_host_network);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						hidden={node.no_host_network}
						label="Host Network NAT"
						help="Enable NAT to on the host network."
						checked={node.host_nat}
						onToggle={(): void => {
							this.set('host_nat', !node.host_nat);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Node Port Network"
						help="Enable node port networking to allow instances to use node ports."
						checked={!node.no_node_port_network}
						onToggle={(): void => {
							this.set('no_node_port_network', !node.no_node_port_network);
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={node.network_mode !== 'oracle' &&
							node.network_mode6 !== 'oracle'}
						label="Oracle Cloud User OCID"
						help="User OCID for Oracle Cloud API authentication."
						type="text"
						placeholder="Enter user OCID"
						value={node.oracle_user}
						onChange={(val): void => {
							this.set('oracle_user', val);
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={node.network_mode !== 'oracle' &&
							node.network_mode6 !== 'oracle'}
						label="Oracle Cloud User Tenancy"
						help="Tenancy OCID for Oracle Cloud API authentication."
						type="text"
						placeholder="Enter tenancy OCID"
						value={node.oracle_tenancy}
						onChange={(val): void => {
							this.set('oracle_tenancy', val);
						}}
					/>
					<PageTextArea
						disabled={this.state.disabled}
						hidden={node.network_mode !== 'oracle' &&
							node.network_mode6 !== 'oracle'}
						label="Oracle Cloud Public Key"
						help="Public key for Oracle Cloud API authentication."
						placeholder="Oracle Cloud public key"
						readOnly={true}
						rows={6}
						value={node.oracle_public_key}
						onChange={(val: string): void => {
							this.set('oracle_public_key', val);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Default instance public IPv4 address"
						help="Enable or disable default option for instance public IPv4 address."
						checked={!node.default_no_public_address}
						onToggle={(): void => {
							this.set('default_no_public_address',
								!node.default_no_public_address);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Default instance public IPv6 address"
						help="Enable or disable default option for instance public IPv6 address."
						checked={!node.default_no_public_address6}
						onToggle={(): void => {
							this.set('default_no_public_address6',
								!node.default_no_public_address6);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Jumbo frames external"
						help="Enable jumbo frames on external interfaces, requires node restart when changed. Node external interfaces must be configured for 9000 MTU. Also requires internal jumbo frames."
						checked={node.jumbo_frames}
						onToggle={(): void => {
							this.set('jumbo_frames', !node.jumbo_frames);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Jumbo frames internal"
						help="Enable jumbo frames on internal interfaces, requires node restart when changed. Node interal interfaces must be configured for 9000 MTU."
						checked={node.jumbo_frames || node.jumbo_frames_internal}
						onToggle={(): void => {
							this.set('jumbo_frames_internal', !node.jumbo_frames_internal);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Instance iSCSI support"
						help="Enable iSCSI disk support for instances."
						checked={node.iscsi}
						hidden={!node.iscsi}
						onToggle={(): void => {
							this.set('iscsi', !node.iscsi);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="PCI Passthough"
						help="Enable PCI passthrough support for instances."
						checked={node.pci_passthrough}
						onToggle={(): void => {
							this.set('pci_passthrough', !node.pci_passthrough);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="USB Passthough"
						help="Enable USB passthrough support for instances."
						checked={node.usb_passthrough}
						onToggle={(): void => {
							this.set('usb_passthrough', !node.usb_passthrough);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="HugePages"
						help="Static hugepages provide a sector of the system memory to be dedicated for hugepages. This memory will be used for instances allowing higher memory performance and preventing the host system from disturbing memory dedicated for virtual instances. This option should always be used on production systems. The hugepages size must be set with the option below or manually with sysctl. Enabling this option while instances are running is likely to crash the system."
						checked={node.hugepages}
						onToggle={(): void => {
							this.set('hugepages', !node.hugepages);
						}}
					/>
					<PageNumInput
						label="HugePages Size"
						help="Size of hugepages space in megabytes. Set this option to the size of memory that will be dedicated for virtual instances. It is recommended to leave 4GB of memory for the host system. Set to 0 if the hugepages size is being manually configured."
						min={0}
						minorStepSize={0}
						stepSize={1024}
						majorStepSize={1024}
						disabled={this.state.disabled}
						hidden={!node.hugepages}
						selectAllOnFocus={true}
						onChange={(val: number): void => {
							this.set('hugepages_size', val);
						}}
						value={node.hugepages_size}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Firewall"
						help="Configure firewall on node. Incorrectly configuring the firewall can block access to the node."
						checked={node.firewall}
						onToggle={(): void => {
							this.toggleFirewall();
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Desktop GUI"
						help="Enable support for desktop GUI display for instances. Requires Xorg or Wayland session to be running."
						checked={node.gui}
						onToggle={(): void => {
							this.set('gui', !node.gui);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Desktop GUI Mode"
						help="Enable support for desktop GUI display for instances. Requires Xorg or Wayland session to be running."
						checked={node.gui}
						onToggle={(): void => {
							this.set('gui', !node.gui);
						}}
					/>
					<PageSelect
						hidden={!node.gui}
						disabled={this.state.disabled}
						label="Desktop GUI Mode"
						help="Desktop GUI display mode. SDL is recommended for better compatibility."
						value={node.gui_mode}
						onChange={(val): void => {
							this.set('gui_mode', val);
						}}
					>
						<option value="sdl">SDL</option>
						<option value="gtk">GTK</option>
					</PageSelect>
					<PageInput
						disabled={this.state.disabled}
						hidden={!node.gui}
						label="Desktop GUI User"
						help="Username of user to open desktop GUI window."
						type="text"
						placeholder="Enter username"
						value={node.gui_user}
						onChange={(val): void => {
							this.set('gui_user', val);
						}}
					/>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.node.id || 'None',
							},
							{
								label: 'Version',
								value: node.software_version || 'Unknown',
							},
							{
								valueClass: active ? '' : 'bp5-text-intent-danger',
								label: 'Timestamp',
								value: MiscUtils.formatDate(
									this.props.node.timestamp) || 'Inactive',
							},
							{
								label: 'CPU Units',
								value: (this.props.node.cpu_units ||
									'Unknown').toString(),
							},
							{
								label: 'CPU Units Reserved',
								value: (this.props.node.cpu_units_res || 0).toString(),
							},
							{
								label: 'Memory Units',
								value: (this.props.node.memory_units ||
									'Unknown').toString(),
							},
							{
								label: 'Memory Units Reserved',
								value: (this.props.node.memory_units_res || 0).toString(),
							},
							{
								label: 'Default Interface',
								value: this.props.node.default_interface || 'Unknown',
							},
							{
								label: 'Hostname',
								value: node.hostname || 'Unknown',
							},
							{
								label: 'Private IPv4',
								value: privateIps,
								copy: true,
							},
							{
								label: 'Public IPv4',
								value: publicIps,
								copy: true,
							},
							{
								label: 'Public IPv6',
								value: publicIps6,
								copy: true,
							},
							{
								label: 'Requests',
								value: this.props.node.requests_min + '/min',
							},
						]}
						bars={resourceBars}
					/>
					<PageSelect
						hidden={types.indexOf('hypervisor') === -1}
						disabled={this.state.disabled}
						label="Hypervisor Mode"
						help="Hypervisor mode, select KVM if CPU has hardware virtualization support."
						value={node.hypervisor}
						onChange={(val): void => {
							this.set('hypervisor', val);
						}}
					>
						<option value="qemu">QEMU</option>
						<option value="kvm">KVM</option>
					</PageSelect>
					<PageSelect
						hidden={types.indexOf('hypervisor') === -1}
						disabled={this.state.disabled}
						label="Hypervisor VGA Type"
						help={<div>
							Type of VGA card to emulate. Virtio provides the best performance.
							VMware provides better performance then standard. Virtio is
							required for UEFI guests.
							<ul>
								<li>Standard = --vga=std</li>
								<li>VMware = --vga=vmware</li>
								<li>Virtio = --display=virtio-vga</li>
								<li>Virtio GPU PCI = --display=virtio-gpu-pci</li>
								<li>Virtio VGA OpenGL = --display=virtio-vga-gl</li>
								<li>Virtio GPU OpenGL = --display=virtio-gpu-gl</li>
								<li>Virtio GPU Vulkan = --display=virtio-gpu-gl,venus=true</li>
								<li>Virtio GPU PCI OpenGL = --display=virtio-gpu-gl-pci</li>
								<li>Virtio GPU PCI Vulkan = --display=virtio-gpu-gl-pci,venus=true</li>
							</ul>
						</div>}
						value={node.vga}
						onChange={(val): void => {
							this.set('vga', val);
						}}
					>
						<option value="Std">Standard</option>
						<option value="VMware">Virtio</option>
						<option value="virtio">Virtio</option>
						<option value="virtio_pci">Virtio GPU PCI</option>
						<option value="virtio_vga_gl">Virtio VGA OpenGL</option>
						<option value="virtio_gl">Virtio GPU OpenGL</option>
						<option value="virtio_gl_vulkan">Virtio GPU Vulkan</option>
						<option value="virtio_pci_gl">Virtio GPU PCI OpenGL</option>
						<option value="virtio_pci_gl_vulkan">Virtio GPU PCI Vulkan</option>
					</PageSelect>
					<PageSelect
						hidden={types.indexOf('hypervisor') === -1 ||
							!NodeTypes.RenderModes.has(node.vga)}
						disabled={this.state.disabled || !hasRenders}
						label="Hypervisor EGL Render"
						help="Graphics card to use for EGL rendering."
						value={node.vga_render}
						onChange={(val): void => {
							this.set('vga_render', val);
						}}
					>
						{rendersSelect}
					</PageSelect>
					<label
						className="bp5-label"
						style={css.label}
					>
						Instance Passthrough Disks
						<Help
							title="Instance Direct Disks"
							content="Disk devices available to instances for passthrough."
						/>
						<div>
							{availableDrives}
						</div>
					</label>
					<PageSelectButton
						label="Add Disk"
						value={this.state.addDrive}
						disabled={!availableDrivesSelect.length || this.state.disabled}
						buttonClass="bp5-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								addDrive: val,
							});
						}}
						onSubmit={this.onAddDrive}
					>
						{availableDrivesSelect}
					</PageSelectButton>
					<label className="bp5-label">
						Network Roles
						<Help
							title="Network Roles"
							content="Network roles that will be matched with firewall rules. Network roles are case-sensitive. Only firewall roles without an organization will match."
						/>
						<div>
							{networkRoles}
						</div>
					</label>
					<PageInputButton
						disabled={this.state.disabled}
						buttonClass="bp5-intent-success bp5-icon-add"
						label="Add"
						type="text"
						placeholder="Add role"
						value={this.state.addNetworkRole}
						onChange={(val): void => {
							this.setState({
								...this.state,
								addNetworkRole: val,
							});
						}}
						onSubmit={this.onAddNetworkRole}
					/>
					<label
						className="bp5-label"
						style={css.label}
						hidden={node.protocol === 'http'}
					>
						Certificates
						<Help
							title="Certificates"
							content="The certificates to use for this nodes web server. The certificates must be valid for all the domains that this node provides access to. This includes the management domain and any service domains."
						/>
						<div>
							{certificates}
						</div>
					</label>
					<PageSelectButton
						hidden={node.protocol === 'http'}
						label="Add Certificate"
						value={this.state.addCert}
						disabled={this.state.disabled || !hasCertificates}
						buttonClass="bp5-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								addCert: val,
							});
						}}
						onSubmit={this.onAddCert}
					>
						{certificatesSelect}
					</PageSelectButton>
					<PageInputSwitch
						disabled={this.state.disabled}
						label="Forwarded for header"
						help="Enable when using a load balancer. This header value will be used to get the users IP address. It is important to only enable this when a load balancer is used. If it is enabled without a load balancer users can spoof their IP address by providing a value for the header that will not be overwritten by a load balancer. Additionally the nodes firewall should be configured to only accept requests from the load balancer to prevent requests being sent directly to the node bypassing the load balancer."
						type="text"
						placeholder="Forwarded for header"
						value={node.forwarded_for_header}
						checked={this.state.forwardedChecked}
						defaultValue="X-Forwarded-For"
						onChange={(state: boolean, val: string): void => {
							let nde: NodeTypes.Node;

							if (this.state.changed) {
								nde = {
									...this.state.node,
								};
							} else {
								nde = {
									...this.props.node,
								};
							}

							nde.forwarded_for_header = val;

							this.setState({
								...this.state,
								changed: true,
								forwardedChecked: state,
								node: nde,
							});
						}}
					/>
					<PageInputSwitch
						label="Forwarded proto header"
						help="Enable when using a load balancer. This header value will be used to get the users protocol. This will redirect users to https when the forwarded protocol is http."
						type="text"
						placeholder="Forwarded proto header"
						value={node.forwarded_proto_header}
						checked={this.state.forwardedProtoChecked}
						defaultValue="X-Forwarded-Proto"
						onChange={(state: boolean, val: string): void => {
							let nde: NodeTypes.Node;

							if (this.state.changed) {
								nde = {
									...this.state.node,
								};
							} else {
								nde = {
									...this.props.node,
								};
							}

							nde.forwarded_proto_header = val;

							this.setState({
								...this.state,
								changed: true,
								forwardedProtoChecked: state,
								node: nde,
							});
						}}
					/>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.node}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						forwardedChecked: false,
						forwardedProtoChecked: false,
						node: null,
					});
				}}
				onSave={this.onSave}
			>
				<NodeDeploy
					disabled={this.state.disabled || this.state.changed}
					node={this.props.node}
					datacenters={this.props.datacenters}
					zones={this.props.zones}
					blocks={this.props.blocks}
				/>
			</PageSave>
		</td>;
	}
}
