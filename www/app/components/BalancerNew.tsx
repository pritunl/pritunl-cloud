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
import PageCreate from './PageCreate';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import PageSelectButton from "./PageSelectButton";
import PageSwitch from "./PageSwitch";
import PageTextArea from "./PageTextArea";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	certificates: CertificateTypes.CertificatesRo;
	datacenters: DatacenterTypes.DatacentersRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	balancer: BalancerTypes.Balancer;
	addCert: string;
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
	rules: {
		marginBottom: '15px',
	} as React.CSSProperties,
};

export default class BalancerNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			addCert: null,
			balancer: {
				name: 'new-balancer',
			},
		};
	}

	set(name: string, val: any): void {
		let balancer: any = {
			...this.state.balancer,
		};

		balancer[name] = val;

		this.setState({
			...this.state,
			changed: true,
			balancer: balancer,
		});
	}

	onAddBackend = (): void => {
		let balancer: BalancerTypes.Balancer;

		balancer = {
			...this.state.balancer,
		};

		let backends = [
			...(balancer.backends || []),
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

		balancer = {
			...this.state.balancer,
		};

		let backends = [
			...(balancer.backends || []),
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

		balancer = {
			...this.state.balancer,
		};

		let backends = [
			...(balancer.backends || []),
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

		balancer = {
			...this.state.balancer,
		};

		if (!this.state.addCert && !this.props.certificates.length) {
			return;
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

		balancer = {
			...this.state.balancer,
		};

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

		balancer = {
			...this.state.balancer,
		};

		let domains = [
			...(balancer.domains || []),
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

		balancer = {
			...this.state.balancer,
		};

		let domains = [
			...(balancer.domains || []),
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

		balancer = {
			...this.state.balancer,
		};

		let domains = [
			...(balancer.domains || []),
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

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let balancer: any = {
			...this.state.balancer,
		};

		BalancerActions.create(balancer).then((): void => {
			this.setState({
				...this.state,
				message: 'Balancer created successfully',
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
		let balancer: BalancerTypes.Balancer = this.state.balancer;

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
		(balancer.domains || []).forEach((domn, index) => {
			domains.push(
				<BalancerDomain
					key={index}
					domain={domn}
					onChange={(state: BalancerTypes.Domain): void => {
						this.onChangeDomain(index, state);
					}}
					onRemove={(): void => {
						this.onRemoveDomain(index);
					}}
				/>,
			);
		})

		let backends: JSX.Element[] = [];
		(balancer.backends || []).forEach((backend, index) => {
			backends.push(
				<BalancerBackend
					key={index}
					backend={backend}
					onChange={(state: BalancerTypes.Backend): void => {
						this.onChangeBackend(index, state);
					}}
					onRemove={(): void => {
						this.onRemoveBackend(index);
					}}
				/>,
			);
		})

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

		return <div
			className="bp5-card bp5-row"
			style={css.row}
		>
			<td
				className="bp5-cell"
				colSpan={3}
				style={css.card}
			>
				<div className="layout horizontal wrap">
					<div style={css.group}>
						<div style={css.buttons}>
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
				<PageCreate
					style={css.save}
					hidden={!this.state.balancer}
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
