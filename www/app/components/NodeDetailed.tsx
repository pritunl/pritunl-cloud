/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as NodeTypes from '../types/NodeTypes';
import * as CertificateTypes from '../types/CertificateTypes';
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as ZoneTypes from '../types/ZoneTypes';
import * as NodeActions from '../actions/NodeActions';
import * as MiscUtils from '../utils/MiscUtils';
import CertificatesStore from '../stores/CertificatesStore';
import PageInput from './PageInput';
import PageSwitch from './PageSwitch';
import PageInputSwitch from './PageInputSwitch';
import PageSelect from './PageSelect';
import PageSelectButton from './PageSelectButton';
import PageInputButton from './PageInputButton';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';

interface Props {
	node: NodeTypes.NodeRo;
	certificates: CertificateTypes.CertificatesRo;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
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
	addCert: string;
	addNetworkRole: string;
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
		minWidth: '90px',
		flex: '0 1 auto',
	} as React.CSSProperties,
	port: {
		minWidth: '120px',
		flex: '1',
	} as React.CSSProperties,
	select: {
		margin: '7px 0px 0px 6px',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		height: '20px',
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
			addCert: null,
			addNetworkRole: null,
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

		if (!this.state.addExternalIface &&
			!this.props.node.available_interfaces.length) {
			return;
		}

		let certId = this.state.addExternalIface ||
			this.props.node.available_interfaces[0];

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

		if (ifaces.indexOf(certId) === -1) {
			ifaces.push(certId);
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

		if (!this.state.addInternalIface &&
				!this.props.node.available_interfaces.length) {
			return;
		}

		let certId = this.state.addInternalIface ||
			this.props.node.available_interfaces[0];

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

		if (ifaces.indexOf(certId) === -1) {
			ifaces.push(certId);
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

		let certId = this.state.addCert || this.props.certificates[0].id;

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

		let externalIfaces: JSX.Element[] = [];
		for (let iface of (node.external_interfaces || [])) {
			externalIfaces.push(
				<div
					className="pt-tag pt-tag-removable pt-intent-primary"
					style={css.item}
					key={iface}
				>
					{iface}
					<button
						disabled={this.state.disabled}
						className="pt-tag-remove"
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
					className="pt-tag pt-tag-removable pt-intent-primary"
					style={css.item}
					key={iface}
				>
					{iface}
					<button
						disabled={this.state.disabled}
						className="pt-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveInternalIface(iface);
						}}
					/>
				</div>,
			);
		}

		let externalIfacesSelect: JSX.Element[] = [];
		let internalIfacesSelect: JSX.Element[] = [];
		for (let iface of (this.props.node.available_interfaces || [])) {
			externalIfacesSelect.push(
				<option key={iface} value={iface}>
					{iface}
				</option>,
			);
			internalIfacesSelect.push(
				<option key={iface} value={iface}>
					{iface}
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
					className="pt-tag pt-tag-removable pt-intent-primary"
					style={css.item}
					key={cert.id}
				>
					{cert.name}
					<button
						disabled={this.state.disabled}
						className="pt-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveCert(cert.id);
						}}
					/>
				</div>,
			);
		}

		let certificatesSelect: JSX.Element[] = [];
		if (this.props.certificates.length) {
			for (let certificate of this.props.certificates) {
				certificatesSelect.push(
					<option key={certificate.id} value={certificate.id}>
						{certificate.name}
					</option>,
				);
			}
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

		let networkRoles: JSX.Element[] = [];
		for (let networkRole of (node.network_roles || [])) {
			networkRoles.push(
				<div
					className="pt-tag pt-tag-removable pt-intent-primary"
					style={css.role}
					key={networkRole}
				>
					{networkRole}
					<button
						className="pt-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveNetworkRole(networkRole);
						}}
					/>
				</div>,
			);
		}

		return <td
			className="pt-cell"
			colSpan={4}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal"
						style={css.buttons}
						onClick={(evt): void => {
							let target = evt.target as HTMLElement;

							if (target.className.indexOf('open-ignore') !== -1) {
								return;
							}

							this.props.onClose();
						}}
					>
						<div className="flex"/>
						<ConfirmButton
							className="pt-minimal pt-intent-danger pt-icon-trash open-ignore"
							progressClassName="pt-intent-danger"
							confirmMsg="Confirm node remove"
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
						label="Hypervisor"
						help="Run instances with hypervisor on this node."
						checked={types.indexOf('hypervisor') !== -1}
						onToggle={(): void => {
							this.toggleType('hypervisor');
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={types.indexOf('admin') === -1 ||
							types.indexOf('user') === -1}
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
						hidden={types.indexOf('admin') === -1 ||
							types.indexOf('user') === -1}
						label="User Domain"
						help="Domain that will be used to access the user interface."
						type="text"
						placeholder="Enter user domain"
						value={node.user_domain}
						onChange={(val): void => {
							this.set('user_domain', val);
						}}
					/>
					<label className="pt-label" style={css.label}>
						Protocol and Port
						<div className="pt-control-group" style={css.inputGroup}>
							<div className="pt-select" style={css.protocol}>
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
								className="pt-input"
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
						help="Node zone, cannot be changed once set."
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
					<label
						className="pt-label"
						style={css.label}
						hidden={node.protocol === 'http'}
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
						hidden={node.protocol === 'http'}
						label="Add Interface"
						value={this.state.addCert}
						disabled={!externalIfacesSelect.length || this.state.disabled}
						buttonClass="pt-intent-success"
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
						className="pt-label"
						style={css.label}
						hidden={node.protocol === 'http'}
					>
						Internal Interfaces
						<Help
							title="Internal Interfaces"
							content="Internal interfaces for instance private VPC interface, must be a bridge interface. Leave blank for to use external interface."
						/>
						<div>
							{internalIfaces}
						</div>
					</label>
					<PageSelectButton
						hidden={node.protocol === 'http'}
						label="Add Interface"
						value={this.state.addCert}
						disabled={!internalIfacesSelect.length || this.state.disabled}
						buttonClass="pt-intent-success"
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
					<PageInput
						disabled={this.state.disabled}
						label="External Interface"
						help="External interface for instance public interface, must be a bridge interface. Leave blank for automatic configuration."
						type="text"
						placeholder="Automatic"
						value={node.external_interface}
						onChange={(val): void => {
							this.set('external_interface', val);
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						label="Internal Interface"
						help="Internal interface for instance private VPC interface, must be a bridge interface. Leave blank for to use external interface."
						type="text"
						placeholder="Automatic"
						value={node.internal_interface}
						onChange={(val): void => {
							this.set('internal_interface', val);
						}}
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
					<PageSwitch
						disabled={this.state.disabled}
						label="Firewall"
						help="Configure firewall on node. Incorrectly configuring the firewall can block access to the node."
						checked={node.firewall}
						onToggle={(): void => {
							this.toggleFirewall();
						}}
					/>
					<label className="pt-label">
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
						buttonClass="pt-intent-success pt-icon-add"
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
								valueClass: active ? '' : 'pt-text-intent-danger',
								label: 'Timestamp',
								value: MiscUtils.formatDate(
									this.props.node.timestamp) || 'Inactive',
							},
							{
								label: 'CPU Units',
								value: (this.props.node.cpu_units || 'Unknown').toString(),
							},
							{
								label: 'Memory Units',
								value: (this.props.node.memory_units || 'Unknown').toString(),
							},
							{
								label: 'Public IPv4',
								value: publicIps,
							},
							{
								label: 'Public IPv6',
								value: publicIps6,
							},
							{
								label: 'Requests',
								value: this.props.node.requests_min + '/min',
							},
						]}
						bars={[
							{
								progressClass: 'pt-no-stripes pt-intent-primary',
								label: 'Memory',
								value: this.props.node.memory,
							},
							{
								progressClass: 'pt-no-stripes pt-intent-success',
								label: 'Load1',
								value: this.props.node.load1,
							},
							{
								progressClass: 'pt-no-stripes pt-intent-warning',
								label: 'Load5',
								value: this.props.node.load5,
							},
							{
								progressClass: 'pt-no-stripes pt-intent-danger',
								label: 'Load15',
								value: this.props.node.load15,
							},
						]}
					/>
					<label
						className="pt-label"
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
						disabled={!this.props.certificates.length || this.state.disabled}
						buttonClass="pt-intent-success"
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
			/>
		</td>;
	}
}
