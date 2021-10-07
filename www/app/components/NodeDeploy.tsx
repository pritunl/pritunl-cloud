/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as NodeTypes from '../types/NodeTypes';
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as ZoneTypes from '../types/ZoneTypes';
import * as NodeActions from '../actions/NodeActions';
import * as BlockTypes from '../types/BlockTypes';
import * as MiscUtils from '../utils/MiscUtils';
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
	datacenter: string;
	zone: string;
	internalIface: string;
	network: string;
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
		fontSize: '12px',
		fontFamily: '"Lucida Console", Monaco, monospace',
	} as React.CSSProperties,
};

export default class NodeDeploy extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			message: '',
			datacenter: '',
			zone: '',
			internalIface: '',
			network: '',
			popover: false,
		};
	}

	ifaces(): string[] {
		let node = this.props.node;

		let zoneId = node.zone;
		if (this.state.zone) {
			zoneId = this.state.zone;
		}

		let vxlan = false;
		for (let zne of this.props.zones) {
			if (zne.id === zoneId) {
				if (zne.network_mode === 'vxlan_vlan') {
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

	onSave = (): void => {
		let internalIface = this.state.internalIface;
		if (!internalIface) {
			let ifaces = this.ifaces();
			if (ifaces.length) {
				internalIface = ifaces[0];
			}
		}

		let data: NodeTypes.NodeInit = {
			zone: this.state.zone,
			internal_interface: internalIface,
			host_network: this.state.network,
		};

		NodeActions.init(this.props.node.id, data).then((): void => {
			// this.setState({
			// 	...this.state,
			// 	message: 'Your changes have been saved',
			// 	disabled: false,
			// });
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

	render(): JSX.Element {
		let popoverElem: JSX.Element;

		if (this.state.popover) {
			let callout = 'Initialize Node. Selected zone must have VXLAN network mode.';
			let errorMsg = '';
			let errorMsgElem: JSX.Element;

			if (errorMsg) {
				errorMsgElem = <div className="bp3-dialog-body">
					<div
						className="bp3-callout bp3-intent-danger bp3-icon-ban-circle"
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
			let internalIfacesSelect: JSX.Element[] = [];
			for (let iface of (availableIfaces || [])) {
				internalIfacesSelect.push(
					<option key={iface} value={iface}>
						{iface}
					</option>,
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
				<div className="bp3-dialog-body" hidden={!!errorMsgElem}>
					<div
						className="bp3-callout bp3-intent-primary bp3-icon-info-sign"
						style={css.callout}
					>
						{callout}
					</div>
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
						disabled={this.state.disabled || !internalIfacesSelect.length}
						label="Network Interface"
						help="Network interface for instance private VPC interface. This interface will be used to send VPC traffic between instances located on multiple nodes. For single node clusters this will not be used."
						value={this.state.internalIface}
						onChange={(val): void => {
							this.setState({
								...this.state,
								internalIface: val,
							});
						}}
					>
						{internalIfacesSelect}
					</PageSelect>
					<PageInput
						disabled={this.state.disabled}
						label="Host IPv4 Network"
						help="Host IPv4 network that is configured on the host to provide networking between the host and the instances. Each node must have a unique host network."
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
					<PageInput
						disabled={this.state.disabled}
						label="Public IPv6 Network"
						help="Public IPv6 network that is routed to this node. Copy the network and prefix, must be at least /64 prefix."
						type="text"
						placeholder="Enter network"
						value={this.state.network6}
						onChange={(val): void => {
							this.setState({
								...this.state,
								network6: val,
							});
						}}
					/>
				</div>
				<div className="bp3-dialog-footer">
					<div className="bp3-dialog-footer-actions">
						<button
							className="bp3-button bp3-icon-cloud-upload bp3-intent-primary"
							type="button"
							onClick={this.onSave}
						>
							Initialize Node
						</button>
						<button
							className="bp3-button"
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									popover: !this.state.popover,
								});
							}}
						>Close</button>
					</div>
				</div>
			</Blueprint.Dialog>;
		}

		return <div hidden={this.props.hidden} style={css.box}>
			<button
				className="bp3-button bp3-icon-cloud-upload bp3-intent-primary"
				style={css.button}
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
