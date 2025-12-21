/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as NodeTypes from '../types/NodeTypes';
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as ZoneTypes from '../types/ZoneTypes';
import * as NodeActions from '../actions/NodeActions';
import * as BlockTypes from '../types/BlockTypes';
import * as MiscUtils from '../utils/MiscUtils';
import * as Theme from '../Theme';
import Help from './Help';
import PageInput from './PageInput';
import PageInputButton from './PageInputButton';
import PageSwitch from './PageSwitch';
import PageSelect from './PageSelect';
import CertificatesStore from "../stores/CertificatesStore";
import NodeBlock from "./NodeBlock";

interface Props {
	hidden?: boolean;
	disabled?: boolean;
	node: NodeTypes.NodeRo;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
	blocks: BlockTypes.BlocksRo;
}

interface State {
	disabled: boolean;
	message: string;
	provider: string;
	datacenter: string;
	zone: string;
	firewall: boolean;
	internalIface: string;
	externalIface: string;
	network: string;
	gateway: string;
	netmask: string;
	subnets: string[];
	addSubnet: string,
	popover: boolean;
}

const css = {
	box: {
	} as React.CSSProperties,
	button: {
		marginRight: '10px',
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	callout: {
		marginBottom: '15px',
	} as React.CSSProperties,
	popover: {
		width: '230px',
	} as React.CSSProperties,
	popoverTarget: {
		top: '9px',
		left: '18px',
	} as React.CSSProperties,
	dialog: {
		maxWidth: '480px',
		margin: '30px 20px',
	} as React.CSSProperties,
	textarea: {
		width: '100%',
		resize: 'none',
		fontSize: Theme.monospaceSize,
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
};

export default class NodeDeploy extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			message: '',
			provider: '',
			datacenter: '',
			zone: '',
			firewall: true,
			internalIface: '',
			externalIface: '',
			network: '',
			gateway: '',
			netmask: '',
			subnets: [],
			addSubnet: '',
			popover: false,
		};
	}

	ifaces(): NodeTypes.Interface[] {
		let node = this.props.node;

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
			return NodeTypes.GetAllIfaces(node);
		} else {
			return NodeTypes.GetAllIfaces(node);
		}
	}

	onSave = (): void => {
		let internalIface = this.state.internalIface;
		if (!internalIface) {
			let ifaces = this.ifaces();
			if (ifaces.length) {
				internalIface = ifaces[0]?.name;
			}
		}

		let externalIface = this.state.externalIface;
		if (!externalIface) {
			let ifaces = this.ifaces();
			if (ifaces.length) {
				externalIface = ifaces[0]?.name;
			}
		}

		let data: NodeTypes.NodeInit = {
			provider: this.state.provider || 'other',
			zone: this.props.node.zone ? this.props.node.zone :
				this.state.zone,
			firewall: this.state.firewall,
			internal_interface: internalIface,
			external_interface: externalIface,
			host_network: this.state.network,
			block_gateway: this.state.gateway,
			block_netmask: this.state.netmask,
			block_subnets: this.state.subnets,
		};

		NodeActions.init(this.props.node.id, data).then((): void => {
			this.setState({
				...this.state,
				popover: !this.state.popover,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	onAddSubnet = (): void => {
		if (!this.state.addSubnet) {
			return;
		}

		let subnets = [
			...this.state.subnets,
		];

		let addSubnet = this.state.addSubnet.trim();
		if (subnets.indexOf(addSubnet) === -1) {
			subnets.push(addSubnet);
		}

		subnets.sort();

		this.setState({
			...this.state,
			subnets: subnets,
			addSubnet: '',
		});
	}

	onRemoveSubnet = (subnet: string): void => {
		let subnets = [
			...(this.state.subnets || []),
		];

		let i = subnets.indexOf(subnet);
		if (i === -1) {
			return;
		}

		subnets.splice(i, 1);

		this.setState({
			...this.state,
			subnets: subnets,
		});
	}

	render(): JSX.Element {
		let popoverElem: JSX.Element;

		if (this.state.popover) {
			let callout = 'Initialize node, select the hosts public network interface.';
			let errorMsg = '';
			let errorMsgElem: JSX.Element;

			if (errorMsg) {
				errorMsgElem = <div className="bp5-dialog-body">
					<div
						className="bp5-callout bp5-intent-danger bp5-icon-ban-circle"
						style={css.callout}
					>
						{errorMsg}
					</div>
				</div>;
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

			let availableIfaces = this.ifaces();
			let ifacesSelect: JSX.Element[] = [];
			for (let iface of (availableIfaces || [])) {
				ifacesSelect.push(
					<option key={iface.name} value={iface.name}>
						{iface.name + (iface.address ? (" (" + iface.address + ")") : "")}
					</option>,
				);
			}

			let subnets: JSX.Element[] = [];
			for (let subnet of (this.state.subnets || [])) {
				subnets.push(
					<div
						className="bp5-tag bp5-tag-removable bp5-intent-primary"
						style={css.item}
						key={subnet}
					>
						{subnet}
						<button
							className="bp5-tag-remove"
							disabled={this.state.disabled}
							onMouseUp={(): void => {
								this.onRemoveSubnet(subnet);
							}}
						/>
					</div>,
				);
			}

			popoverElem = <Blueprint.Dialog
				title="Initialize Node"
				style={css.dialog}
				isOpen={this.state.popover}
				usePortal={true}
				portalContainer={document.body}
				onClose={(): void => {
					this.setState({
						...this.state,
						popover: false,
					});
				}}
			>
				{errorMsgElem}
				<div className="bp5-dialog-body" hidden={!!errorMsgElem}>
					<div
						className="bp5-callout bp5-intent-primary bp5-icon-info-sign"
						style={css.callout}
					>
						{callout}
					</div>
					<PageSelect
						disabled={this.state.disabled}
						label="Provider"
						help="Bare metal hosting provider."
						value={this.state.provider}
						onChange={(val): void => {
							this.setState({
								...this.state,
								provider: val,
							});
						}}
					>
						<option key="other" value="other">Other</option>
						<option key="vultr" value="vultr">Vultr</option>
						<option key="phoenixnap" value="phoenixnap">phoenixNAP</option>
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled || !hasDatacenters}
						hidden={!!this.props.node.zone}
						label="Datacenter"
						help="Node datacenter, cannot be changed once set."
						value={this.state.datacenter}
						onChange={(val): void => {
							this.setState({
								...this.state,
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
						help="Node zone, cannot be changed once set."
						value={this.props.node.zone ? this.props.node.zone :
							this.state.zone}
						onChange={(val): void => {
							this.setState({
								...this.state,
								zone: val,
							});
						}}
					>
						{zonesSelect}
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled || !ifacesSelect.length}
						hidden={this.state.provider === 'phoenixnap'}
						label="Public Network Interface"
						help="Network interface for instance public traffic."
						value={this.state.externalIface}
						onChange={(val): void => {
							this.setState({
								...this.state,
								externalIface: val,
							});
						}}
					>
						{ifacesSelect}
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled || !ifacesSelect.length}
						hidden={this.state.provider !== 'phoenixnap'}
						label="Private Network Interface"
						help="Network interface for instance private VPC interface."
						value={this.state.internalIface}
						onChange={(val): void => {
							this.setState({
								...this.state,
								internalIface: val,
							});
						}}
					>
						{ifacesSelect}
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled || !ifacesSelect.length}
						hidden={this.state.provider !== 'phoenixnap'}
						label="Public Network Interface"
						help="Network interface for instance public traffic."
						value={this.state.externalIface}
						onChange={(val): void => {
							this.setState({
								...this.state,
								externalIface: val,
							});
						}}
					>
						{ifacesSelect}
					</PageSelect>
					<PageInput
						disabled={this.state.disabled}
						hidden={this.state.provider !== 'phoenixnap'}
						label="Public Gateway"
						help="Gateway address with prefix for public IP network."
						type="text"
						placeholder="Enter gateway"
						value={this.state.gateway}
						onChange={(val): void => {
							this.setState({
								...this.state,
								gateway: val,
							});
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={true}
						label="Public Netmask"
						help="Netmask of of public IP addresses"
						type="text"
						placeholder="Enter netmask"
						value={this.state.netmask}
						onChange={(val): void => {
							this.setState({
								...this.state,
								netmask: val,
							});
						}}
					/>
					<label
						className="bp5-label"
						hidden={this.state.provider !== 'phoenixnap'}
					>
						IP Addresses
						<Help
							title="Public IP Addresses"
							content="Public IP addresses that are available for instances."
						/>
						<div>
							{subnets}
						</div>
					</label>
					<PageInputButton
						disabled={this.state.disabled}
						hidden={this.state.provider !== 'phoenixnap'}
						buttonClass="bp5-intent-success bp5-icon-add"
						label="Add"
						type="text"
						placeholder="Add addresses"
						value={this.state.addSubnet}
						onChange={(val): void => {
							this.setState({
								...this.state,
								addSubnet: val,
							});
						}}
						onSubmit={this.onAddSubnet}
					/>
					<PageInput
						disabled={this.state.disabled}
						label="Host IPv4 Network"
						help="Host IPv4 network with prefix that is configured on the host to provide networking between the host and the instances. If left blank no host network will be created."
						type="text"
						placeholder="Enter network"
						value={this.state.network}
						onChange={(val): void => {
							this.setState({
								...this.state,
								network: val,
							});
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Node firewall"
						help="Configure a default firewall for the node allowing web and ssh traffic from all addresses. This should always be enabled unless an external firewall has been configured on the host system. The firewall can be modified after from the web console."
						checked={this.state.firewall}
						onToggle={(): void => {
							this.setState({
								...this.state,
								firewall: !this.state.firewall,
							});
						}}
					/>
				</div>
				<div className="bp5-dialog-footer">
					<div className="bp5-dialog-footer-actions">
						<button
							className="bp5-button"
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									popover: !this.state.popover,
								});
							}}
						>Close</button>
						<button
							className="bp5-button bp5-icon-cloud-upload bp5-intent-primary"
							type="button"
							onClick={this.onSave}
						>
							Initialize Node
						</button>
					</div>
				</div>
			</Blueprint.Dialog>;
		}

		return <div hidden={this.props.hidden} style={css.box}>
			<button
				className="bp5-button bp5-icon-cloud-upload bp5-intent-primary"
				style={css.button}
				hidden={true}
				type="button"
				disabled={this.props.disabled}
				onClick={(): void => {
					this.setState({
						...this.state,
						popover: !this.state.popover,
					});
				}}
			>
				Initialize Node
			</button>
			{popoverElem}
		</div>;
	}
}
