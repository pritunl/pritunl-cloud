/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as BalancerTypes from '../types/BalancerTypes';
import * as BalancerActions from '../actions/BalancerActions';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as CertificateTypes from '../types/CertificateTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import BalancerDomain from './BalancerDomain';
import BalancerBackend from './BalancerBackend';
import CompletionStore from '../stores/CompletionStore';
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import PageSelectButton from "./PageSelectButton";
import PageSwitch from "./PageSwitch";
import PageTextArea from "./PageTextArea";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	certificates: CertificateTypes.CertificatesRo;
	datacenters: DatacenterTypes.DatacentersRo;
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
				protocol: 'http',
				hostname: '',
				port: 80,
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

		if (this.state.changed) {
			balancer = {
				...this.state.balancer,
			};
		} else {
			balancer = {
				...this.props.balancer,
			};
		}

		let certId = this.state.addCert;
		if (!certId) {
			for (let certificate of this.props.certificates) {
				if (certificate.organization != balancer.organization) {
					continue;
				}
				certId = certificate.id;
				break;
			}
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
			let cert = CompletionStore.certificate(certId);
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
		if (this.props.certificates) {
			for (let certificate of this.props.certificates) {
				if (certificate.organization != balancer.organization) {
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

		let requests = 0;
		let retries = 0;
		let websockets = 0;
		let states: string[] = [];
		let statesMap: {[index: string]: number} = {};
		let online: string[] = [];
		let unknownHigh: string[] = [];
		let unknownMid: string[] = [];
		let unknownLow: string[] = [];
		let offline: string[] = [];
		let backendsClasses: string[] = [];

		if (this.props.balancer.state && balancer.states) {
			for (let key in balancer.states) {
				if (!balancer.states.hasOwnProperty(key)) {
					continue;
				}
				let state = balancer.states[key];

				requests += state.requests || 0;
				retries += state.retries || 0;
				websockets += state.websockets || 0;

				for (let backend of state.offline) {
					let curState = statesMap[backend];
					if (curState === undefined || curState > 1) {
						statesMap[backend] = 1;
					}
				}
				for (let backend of state.unknown_low) {
					let curState = statesMap[backend];
					if (curState === undefined || curState > 2) {
						statesMap[backend] = 2;
					}
				}
				for (let backend of state.unknown_mid) {
					let curState = statesMap[backend];
					if (curState === undefined || curState > 3) {
						statesMap[backend] = 3;
					}
				}
				for (let backend of state.unknown_high) {
					let curState = statesMap[backend];
					if (curState === undefined || curState > 4) {
						statesMap[backend] = 4;
					}
				}
				for (let backend of state.online) {
					let curState = statesMap[backend];
					if (curState === undefined) {
						statesMap[backend] = 5;
					}
				}
			}

			for (let backend in statesMap) {
				if (!statesMap.hasOwnProperty(backend)) {
					continue;
				}
				let state = statesMap[backend];

				switch (state) {
					case 5:
						online.push(backend);
						break;
					case 4:
						unknownHigh.push(backend);
						break;
					case 3:
						unknownMid.push(backend);
						break;
					case 2:
						unknownLow.push(backend);
						break;
					case 1:
						offline.push(backend);
						break;
				}
			}

			online.sort();
			for (let backend of online) {
				states.push(backend + ' - Online');
				backendsClasses.push('bp5-text-intent-success');
			}
			unknownHigh.sort();
			for (let backend of unknownHigh) {
				states.push(backend + ' - Unknown High');
				backendsClasses.push('bp5-text-intent-warning');
			}
			unknownMid.sort();
			for (let backend of unknownMid) {
				states.push(backend + ' - Unknown Mid');
				backendsClasses.push('bp5-text-intent-warning');
			}
			unknownLow.sort();
			for (let backend of unknownLow) {
				states.push(backend + ' - Unknown Low');
				backendsClasses.push('bp5-text-intent-warning');
			}
			offline.sort();
			for (let backend of offline) {
				states.push(backend + ' - Offline');
				backendsClasses.push('bp5-text-intent-danger');
			}
		}

		if (!states.length) {
			states = ['-'];
		}

		return <td
			className="bp5-cell"
			colSpan={3}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal tab-close bp5-card-header"
						style={css.buttons}
						onClick={(evt): void => {
							if (evt.target instanceof HTMLElement &&
									evt.target.className.indexOf('tab-close') !== -1) {
								this.props.onClose();
							}
						}}
					>
            <div>
              <label
                className="bp5-control bp5-checkbox bp5-checkbox"
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
							dialogLabel="Delete Balancer"
							confirmMsg="Permanently delete this balancer"
							confirmInput={true}
							items={[balancer.name]}
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						disabled={this.state.disabled}
						label="Name"
						help="Name of load balancer"
						type="text"
						placeholder="Enter name"
						value={balancer.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageTextArea
						label="Comment"
						help="Load balancer comment."
						placeholder="Load balancer comment"
						rows={3}
						value={balancer.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Active"
						help="Enable or disable load balancer."
						checked={balancer.state}
						onToggle={(): void => {
							this.set('state', !balancer.state);
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
						label="Datacenter"
						help="Load balancer datacenter."
						value={balancer.datacenter}
						onChange={(val): void => {
							this.set('datacenter', val);
						}}
					>
						{datacentersSelect}
					</PageSelect>
					<label style={css.itemsLabel}>
						External Domains
						<Help
							title="External Domains"
							content="When a request comes into a node the requests host will be used to match the request with the domain of a load balancer. Some internal services will be expecting a specific host such as a web server that serves mutliple websites that is also matching the requests host to one of the mutliple websites. If the internal service is expecting a different host set the host field, otherwise leave it blank. Load balancers that are associated with the same datacenter should not also have the same domains."
						/>
					</label>
					{domains}
					<button
						className="bp5-button bp5-intent-success bp5-icon-add"
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
						className="bp5-button bp5-intent-success bp5-icon-add"
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
							{
								label: 'Requests',
								value: requests + '/min',
							},
							{
								label: 'Retries',
								value: retries + '/min',
							},
							{
								label: 'WebSockets',
								value: websockets,
							},
							{
								label: 'Backends',
								value: states,
								valueClasses: backendsClasses,
							}
						]}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="WebSockets"
						help="Enable or disable WebSocket support on balancer."
						checked={balancer.websockets}
						onToggle={(): void => {
							this.set('websockets', !balancer.websockets);
						}}
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
					<label
						className="bp5-label"
						style={css.label}
					>
						Certificates
						<Help
							title="Certificates"
							content="The certificates to use for this load balancer. The certificates must be valid for all the domains that this load balancer provides access to."
						/>
						<div>
							{certificates}
						</div>
					</label>
					<PageSelectButton
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
					<PageInput
						label="Health Check Path"
						help="Path to check status of backend servers. Path must return 200-299 status code."
						type="text"
						placeholder="Enter path"
						value={balancer.check_path}
						onChange={(val): void => {
							this.set('check_path', val);
						}}
					/>
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
