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
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';

interface Props {
	node: NodeTypes.NodeRo;
	certificates: CertificateTypes.CertificatesRo;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	datacenter: string;
	zone: string;
	changed: boolean;
	message: string;
	node: NodeTypes.Node;
	addCert: string;
	forwardedChecked: boolean;
}

const css = {
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
			addCert: null,
			forwardedChecked: false,
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

	toggleType(typ: string): void {
		let node: NodeTypes.Node = this.state.node || this.props.node;

		let vals = (node.type || '').split('_');

		let i = vals.indexOf(typ);
		if (i === -1) {
			vals.push(typ);
		} else {
			vals.splice(i, 1);
		}

		vals.sort();

		let val = vals.join('_');
		this.set('type', val);
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

		return <td
			className="pt-cell"
			colSpan={4}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div style={css.buttons}>
						<button
							className="pt-button pt-minimal pt-intent-warning pt-icon-chevron-up"
							type="button"
							onClick={(): void => {
								this.props.onClose();
							}}
						/>
						<ConfirmButton
							className="pt-minimal pt-intent-danger pt-icon-cross"
							progressClassName="pt-intent-danger"
							confirmMsg="Confirm node remove"
							disabled={active || this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
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
						label="Admin"
						help="Provides access to the admin console."
						checked={node.type.indexOf('admin') !== -1}
						onToggle={(): void => {
							this.toggleType('admin');
						}}
					/>
					<PageSwitch
						label="User"
						help="Provides access to the user console for SSH certificates."
						checked={node.type.indexOf('user') !== -1}
						onToggle={(): void => {
							this.toggleType('user');
						}}
					/>
					<PageInput
						hidden={node.type.indexOf('_') === -1 ||
							node.type.indexOf('admin') === -1}
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
						hidden={node.type.indexOf('_') === -1 ||
							node.type.indexOf('user') === -1}
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
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: node.id || 'None',
							},
							{
								valueClass: active ? '' : 'pt-text-intent-danger',
								label: 'Timestamp',
								value: MiscUtils.formatDate(node.timestamp) || 'Inactive',
							},
							{
								label: 'Requests',
								value: node.requests_min + '/min',
							},
						]}
						bars={[
							{
								progressClass: 'pt-no-stripes pt-intent-primary',
								label: 'Memory',
								value: node.memory,
							},
							{
								progressClass: 'pt-no-stripes pt-intent-success',
								label: 'Load1',
								value: node.load1,
							},
							{
								progressClass: 'pt-no-stripes pt-intent-warning',
								label: 'Load5',
								value: node.load5,
							},
							{
								progressClass: 'pt-no-stripes pt-intent-danger',
								label: 'Load15',
								value: node.load15,
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
						disabled={!this.props.certificates.length}
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
						node: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}
