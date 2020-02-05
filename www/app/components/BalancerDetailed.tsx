/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as BalancerTypes from '../types/BalancerTypes';
import * as BalancerActions from '../actions/BalancerActions';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as CertificateTypes from '../types/CertificateTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import BalancerDomain from './BalancerDomain';
import BalancerBackend from './BalancerBackend';
import CertificatesStore from '../stores/CertificatesStore';
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	certificates: CertificateTypes.CertificatesRo;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
	balancer: BalancerTypes.BalancerRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	balancer: BalancerTypes.Balancer;
	datacenter: string;
	addCert: string;
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
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	rules: {
		marginBottom: '15px',
	} as React.CSSProperties,
};

export default class BalancerDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			balancer: null,
			datacenter: '',
			addCert: null,
		};
	}

	set(name: string, val: any): void {
		let balancer: any;

		if (this.state.changed) {
			balancer = {
				...this.state.balancer,
			};
		} else {
			balancer = {
				...this.props.balancer,
			};
		}

		balancer[name] = val;

		this.setState({
			...this.state,
			changed: true,
			balancer: balancer,
		});
	}

	onAddBackend = (): void => {
		let balancer: BalancerTypes.Balancer;

		if (this.state.changed) {
			balancer = {
				...this.state.balancer,
			};
		} else {
			balancer = {
				...this.props.balancer,
			};
		}

		let backends = [
			...balancer.backends,
			{
				protocol: 'https',
				hostname: '',
				port: 443,
			},
		];

		balancer.backends = backends;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			balancer: balancer,
		});
	}

	onChangeBackend(i: number, state: BalancerTypes.Backend): void {
		let balancer: BalancerTypes.Balancer;

		if (this.state.changed) {
			balancer = {
				...this.state.balancer,
			};
		} else {
			balancer = {
				...this.props.balancer,
			};
		}

		let backends = [
			...balancer.backends,
		];

		backends[i] = state;

		balancer.backends = backends;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			balancer: balancer,
		});
	}

	onRemoveBackend(i: number): void {
		let balancer: BalancerTypes.Balancer;

		if (this.state.changed) {
			balancer = {
				...this.state.balancer,
			};
		} else {
			balancer = {
				...this.props.balancer,
			};
		}

		let backends = [
			...balancer.backends,
		];

		backends.splice(i, 1);

		balancer.backends = backends;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			balancer: balancer,
		});
	}

	onAddCert = (): void => {
		let balancer: BalancerTypes.Balancer;

		if (!this.state.addCert && !this.props.certificates.length) {
			return;
		}

		let certId = this.state.addCert || this.props.certificates[0].id;

		if (this.state.changed) {
			balancer = {
				...this.state.balancer,
			};
		} else {
			balancer = {
				...this.props.balancer,
			};
		}

		let certificates = [
			...(balancer.certificates || []),
		];

		if (certificates.indexOf(certId) === -1) {
			certificates.push(certId);
		}

		certificates.sort();

		balancer.certificates = certificates;

		this.setState({
			...this.state,
			changed: true,
			balancer: balancer,
		});
	}

	onRemoveCert = (certId: string): void => {
		let balancer: BalancerTypes.Balancer;

		if (this.state.changed) {
			balancer = {
				...this.state.balancer,
			};
		} else {
			balancer = {
				...this.props.balancer,
			};
		}

		let certificates = [
			...(balancer.certificates || []),
		];

		let i = certificates.indexOf(certId);
		if (i === -1) {
			return;
		}

		certificates.splice(i, 1);

		balancer.certificates = certificates;

		this.setState({
			...this.state,
			changed: true,
			balancer: balancer,
		});
	}

	onAddDomain = (): void => {
		let balancer: BalancerTypes.Balancer;

		if (this.state.changed) {
			balancer = {
				...this.state.balancer,
			};
		} else {
			balancer = {
				...this.props.balancer,
			};
		}

		let domains = [
			...balancer.domains,
			{},
		];

		balancer.domains = domains;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			balancer: balancer,
		});
	}

	onChangeDomain(i: number, state: BalancerTypes.Domain): void {
		let balancer: BalancerTypes.Balancer;

		if (this.state.changed) {
			balancer = {
				...this.state.balancer,
			};
		} else {
			balancer = {
				...this.props.balancer,
			};
		}

		let domains = [
			...balancer.domains,
		];

		domains[i] = state;

		balancer.domains = domains;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			balancer: balancer,
		});
	}

	onRemoveDomain(i: number): void {
		let balancer: BalancerTypes.Balancer;

		if (this.state.changed) {
			balancer = {
				...this.state.balancer,
			};
		} else {
			balancer = {
				...this.props.balancer,
			};
		}

		let domains = [
			...balancer.domains,
		];

		domains.splice(i, 1);

		balancer.domains = domains;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			balancer: balancer,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		BalancerActions.commit(this.state.balancer).then((): void => {
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
						balancer: null,
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
		BalancerActions.remove(this.props.balancer.id).then((): void => {
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
		let balancer: BalancerTypes.Balancer = this.state.balancer ||
			this.props.balancer;

		let hasOrganizations = false;
		let organizationsSelect: JSX.Element[] = [];
		if (this.props.organizations.length) {
			organizationsSelect.push(
				<option key="null" value="">
					Select Organization
				</option>,
			);

			for (let organization of this.props.organizations) {
				hasOrganizations = true;

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

		let domains: JSX.Element[] = [];
		for (let i = 0; i < balancer.domains.length; i++) {
			let index = i;

			domains.push(
				<BalancerDomain
					key={index}
					domain={balancer.domains[index]}
					onChange={(state: BalancerTypes.Domain): void => {
						this.onChangeDomain(index, state);
					}}
					onRemove={(): void => {
						this.onRemoveDomain(index);
					}}
				/>,
			);
		}

		let backends: JSX.Element[] = [];
		for (let i = 0; i < balancer.backends.length; i++) {
			let index = i;

			backends.push(
				<BalancerBackend
					key={index}
					backend={balancer.backends[index]}
					onChange={(state: BalancerTypes.Backend): void => {
						this.onChangeBackend(index, state);
					}}
					onRemove={(): void => {
						this.onRemoveBackend(index);
					}}
				/>,
			);
		}

		let certificates: JSX.Element[] = [];
		for (let certId of (balancer.certificates || [])) {
			let cert = CertificatesStore.certificate(certId);
			if (!cert || cert.organization !== balancer.organization) {
				continue;
			}

			certificates.push(
				<div
					className="bp3-tag bp3-tag-removable bp3-intent-primary"
					style={css.item}
					key={cert.id}
				>
					{cert.name}
					<button
						disabled={this.state.disabled}
						className="bp3-tag-remove"
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
				if (!this.props.balancer.zone && zone.datacenter !== datacenter) {
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
			className="bp3-cell"
			colSpan={5}
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
            <div>
              <label
                className="bp3-control bp3-checkbox open-ignore"
                style={css.select}
              >
                <input
                  type="checkbox"
                  className="open-ignore"
                  checked={this.props.selected}
                  onClick={(evt): void => {
										this.props.onSelect(evt.shiftKey);
									}}
                />
                <span className="bp3-control-indicator open-ignore"/>
              </label>
            </div>
						<div className="flex"/>
						<ConfirmButton
							className="bp3-minimal bp3-intent-danger bp3-icon-trash open-ignore"
							style={css.button}
							progressClassName="bp3-intent-danger"
							confirmMsg="Confirm balancer remove"
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of balancer"
						type="text"
						placeholder="Enter name"
						value={balancer.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageSelect
						label="Type"
						help="Load balancer type"
						value={balancer.type}
						onChange={(val): void => {
							this.set('type', val);
						}}
					>
						<option value="http">HTTP</option>
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled || !hasDatacenters}
						hidden={!!this.props.balancer.zone}
						label="Datacenter"
						help="Load balancer datacenter."
						value={this.state.datacenter}
						onChange={(val): void => {
							if (this.state.changed) {
								balancer = {
									...this.state.balancer,
								};
							} else {
								balancer = {
									...this.props.balancer,
								};
							}

							balancer.zone = null;

							this.setState({
								...this.state,
								changed: true,
								balancer: balancer,
								datacenter: val,
							});
						}}
					>
						{datacentersSelect}
					</PageSelect>
					<PageSelect
						disabled={!!this.props.balancer.zone || this.state.disabled ||
						!hasZones}
						label="Zone"
						help="Load balancer zone."
						value={balancer.zone}
						onChange={(val): void => {
							this.set('zone', val);
						}}
					>
						{zonesSelect}
					</PageSelect>
					<label style={css.itemsLabel}>
						External Domains
						<Help
							title="External Domains"
							content="When a request comes into a node the requests host will be used to match the request with the domain of a load balancer. Some internal services will be expecting a specific host such as a web server that serves mutliple websites that is also matching the requests host to one of the mutliple websites. If the internal service is expecting a different host set the host field, otherwise leave it blank. Load balancers that are associated with the same zone should not also have the same domains."
						/>
					</label>
					{domains}
					<button
						className="bp3-button bp3-intent-success bp3-icon-add"
						style={css.itemsAdd}
						type="button"
						onClick={this.onAddDomain}
					>
						Add Domain
					</button>
					<label style={css.itemsLabel}>
						Internal Backends
						<Help
							title="Internal Backends"
							content="After a node receives a request it will be forwarded to the internal servers and the response will be sent back to the user. Multiple internal servers can be added to balance the requests between the servers. If a domain is used with HTTPS the internal server must have a valid certificate. When an IP address is used with HTTPS the internal servers certificate will not be validated."
						/>
					</label>
					{backends}
					<button
						className="bp3-button bp3-intent-success bp3-icon-add"
						style={css.itemsAdd}
						type="button"
						onClick={this.onAddBackend}
					>
						Add Backend
					</button>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.balancer.id || 'Unknown',
							},
						]}
					/>
					<PageSelect
						disabled={this.state.disabled || !hasOrganizations}
						hidden={Constants.user}
						label="Organization"
						help="Organization for balancer, both the organaization and role must match. Select balancer balancer to match balancer network roles."
						value={balancer.organization}
						onChange={(val): void => {
							this.set('organization', val);
						}}
					>
						{organizationsSelect}
					</PageSelect>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.balancer && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						balancer: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}
