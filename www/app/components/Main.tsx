/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ReactRouter from 'react-router-dom';
import * as Theme from '../Theme';
import * as Constants from '../Constants';
import * as SubscriptionTypes from '../types/SubscriptionTypes';
import SubscriptionStore from '../stores/SubscriptionStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import LoadingBar from './LoadingBar';
import Subscription from './Subscription';
import Users from './Users';
import UserDetailed from './UserDetailed';
import Nodes from './Nodes';
import Policies from './Policies';
import Certificates from './Certificates';
import Organizations from './Organizations';
import Datacenters from './Datacenters';
import Zones from './Zones';
import Vpcs from './Vpcs';
import Domains from './Domains';
import Storages from './Storages';
import Images from './Images';
import Disks from './Disks';
import Instances from './Instances';
import Firewalls from './Firewalls';
import Authorities from './Authorities';
import Logs from './Logs';
import Settings from './Settings';
import OrganizationSelect from './OrganizationSelect';
import * as UserActions from '../actions/UserActions';
import * as SessionActions from '../actions/SessionActions';
import * as AuditActions from '../actions/AuditActions';
import * as NodeActions from '../actions/NodeActions';
import * as PolicyActions from '../actions/PolicyActions';
import * as CertificateActions from '../actions/CertificateActions';
import * as OrganizationActions from '../actions/OrganizationActions';
import * as DatacenterActions from '../actions/DatacenterActions';
import * as ZoneActions from '../actions/ZoneActions';
import * as VpcActions from '../actions/VpcActions';
import * as DomainActions from '../actions/DomainActions';
import * as StorageActions from '../actions/StorageActions';
import * as ImageActions from '../actions/ImageActions';
import * as DiskActions from '../actions/DiskActions';
import * as InstanceActions from '../actions/InstanceActions';
import * as FirewallActions from '../actions/FirewallActions';
import * as AuthorityActions from '../actions/AuthorityActions';
import * as LogActions from '../actions/LogActions';
import * as SettingsActions from '../actions/SettingsActions';
import * as SubscriptionActions from '../actions/SubscriptionActions';

interface State {
	subscription: SubscriptionTypes.SubscriptionRo;
	current: string;
	disabled: boolean;
}

const css = {
	card: {
		minWidth: '310px',
		maxWidth: '380px',
		width: 'calc(100% - 20px)',
		margin: '60px auto',
	} as React.CSSProperties,
	nav: {
		overflowX: 'auto',
		overflowY: 'auto',
		userSelect: 'none',
		height: 'auto',
	} as React.CSSProperties,
	navTitle: {
		flexWrap: 'wrap',
		height: 'auto',
	} as React.CSSProperties,
	navGroup: {
		flexWrap: 'wrap',
		height: 'auto',
		padding: '10px 0',
	} as React.CSSProperties,
	link: {
		padding: '0 8px',
		color: 'inherit',
	} as React.CSSProperties,
	sub: {
		color: 'inherit',
	} as React.CSSProperties,
	heading: {
		marginRight: '11px',
		fontSize: '18px',
		fontWeight: 'bold',
	} as React.CSSProperties,
};

export default class Main extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			subscription: SubscriptionStore.subscription,
			current: OrganizationsStore.current,
			disabled: false,
		};
	}

	componentDidMount(): void {
		if (!Constants.user) {
			SubscriptionStore.addChangeListener(this.onChange);
			SubscriptionActions.sync(false);
		} else {
			OrganizationsStore.addChangeListener(this.onChange);
			OrganizationActions.sync();
		}
	}

	componentWillUnmount(): void {
		if (!Constants.user) {
			SubscriptionStore.removeChangeListener(this.onChange);
		} else {
			OrganizationsStore.removeChangeListener(this.onChange);
		}
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			subscription: SubscriptionStore.subscription,
			current: OrganizationsStore.current,
		});
	}

	render(): JSX.Element {
		if ((!Constants.user && !this.state.subscription) ||
				(Constants.user && this.state.current === undefined)) {
			return <div/>;
		}

		if (Constants.user && !this.state.current) {
			return <div>
				<div
					className="pt-callout pt-intent-danger pt-icon-error"
					style={css.card}
				>
					<h4 className="pt-callout-title">No Organization</h4>
					Account does not have access to any organizations
					<button
						className="pt-button pt-minimal pt-icon-log-out"
						onClick={() => {
							window.location.href = '/logout';
						}}
					>Logout</button>
				</div>
			</div>;
		}

		return <ReactRouter.HashRouter>
			<div>
				<nav className="pt-navbar layout horizontal" style={css.nav}>
					<div
						className="pt-navbar-group pt-align-left flex"
						style={css.navTitle}
					>
						<div className="pt-navbar-heading"
							style={css.heading}
						>Pritunl Cloud</div>
						<OrganizationSelect hidden={!Constants.user}/>
					</div>
					<div className="pt-navbar-group pt-align-right" style={css.navGroup}>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-people"
							style={css.link}
							hidden={Constants.user}
							to="/users"
						>
							Users
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-layers"
							style={css.link}
							hidden={Constants.user}
							to="/nodes"
						>
							Nodes
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-filter"
							style={css.link}
							hidden={Constants.user}
							to="/policies"
						>
							Policies
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-endorsed"
							style={css.link}
							hidden={Constants.user}
							to="/certificates"
						>
							Certificates
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-people"
							style={css.link}
							hidden={Constants.user}
							to="/organizations"
						>
							Organizations
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-cloud"
							style={css.link}
							hidden={Constants.user}
							to="/datacenters"
						>
							Datacenters
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-layout-circle"
							style={css.link}
							hidden={Constants.user}
							to="/zones"
						>
							Zones
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-layout-auto"
							style={css.link}
							to="/vpcs"
						>
							VPCs
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-map-marker"
							style={css.link}
							to="/domains"
						>
							Domains
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-database"
							style={css.link}
							hidden={Constants.user}
							to="/storages"
						>
							Storages
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-compressed"
							style={css.link}
							to="/images"
						>
							Images
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-floppy-disk"
							style={css.link}
							to="/disks"
						>
							Disks
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-dashboard"
							style={css.link}
							to="/instances"
						>
							Instances
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-key"
							style={css.link}
							to="/firewalls"
						>
							Firewalls
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-office"
							style={css.link}
							to="/authorities"
						>
							Authorities
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-history"
							style={css.link}
							hidden={Constants.user}
							to="/logs"
						>
							Logs
						</ReactRouter.Link>
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-settings"
							style={css.link}
							hidden={Constants.user}
							to="/settings"
						>
							Settings
						</ReactRouter.Link>
						<ReactRouter.Link
							to="/subscription"
							style={css.sub}
							hidden={Constants.user}
						>
							<button
								className="pt-button pt-minimal pt-icon-credit-card"
								style={css.link}
								onClick={(): void => {
									SubscriptionActions.sync(true);
								}}
							>Subscription</button>
						</ReactRouter.Link>
						<ReactRouter.Route render={(props) => (
							<button
								className="pt-button pt-minimal pt-icon-refresh"
								disabled={this.state.disabled}
								onClick={() => {
									let pathname = props.location.pathname;

									this.setState({
										...this.state,
										disabled: true,
									});

									if (pathname === '/users') {
										UserActions.sync().then((): void => {
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
									} else if (pathname.startsWith('/user/')) {
										UserActions.reload().then((): void => {
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
										SessionActions.reload().then((): void => {
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
										AuditActions.reload().then((): void => {
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
									} else if (pathname === '/nodes') {
										NodeActions.sync().then((): void => {
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
									} else if (pathname === '/policies') {
										SettingsActions.sync();
										PolicyActions.sync().then((): void => {
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
									} else if (pathname === '/certificates') {
										CertificateActions.sync().then((): void => {
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
									} else if (pathname === '/organizations') {
										OrganizationActions.sync().then((): void => {
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
									} else if (pathname === '/datacenters') {
										DatacenterActions.sync().then((): void => {
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
									} else if (pathname === '/zones') {
										ZoneActions.sync().then((): void => {
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
									} else if (pathname === '/vpcs') {
										VpcActions.sync().then((): void => {
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
									} else if (pathname === '/domains') {
										DomainActions.sync().then((): void => {
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
									} else if (pathname === '/storages') {
										StorageActions.sync().then((): void => {
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
									} else if (pathname === '/images') {
										ImageActions.sync().then((): void => {
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
									} else if (pathname === '/disks') {
										DiskActions.sync().then((): void => {
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
									} else if (pathname === '/instances') {
										InstanceActions.sync().then((): void => {
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
									} else if (pathname === '/firewalls') {
										FirewallActions.sync().then((): void => {
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
									} else if (pathname === '/authorities') {
										AuthorityActions.sync().then((): void => {
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
									} else if (pathname === '/logs') {
										LogActions.sync().then((): void => {
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
									} else if (pathname === '/settings') {
										SettingsActions.sync().then((): void => {
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
									} else if (pathname === '/subscription') {
										SubscriptionActions.sync(true).then((): void => {
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
									} else {
										this.setState({
											...this.state,
											disabled: false,
										});
									}
								}}
							>Refresh</button>
						)}/>
						<button
							className="pt-button pt-minimal pt-icon-log-out"
							onClick={() => {
								window.location.href = '/logout';
							}}
						>Logout</button>
						<button
							className="pt-button pt-minimal pt-icon-moon"
							onClick={(): void => {
								Theme.toggle();
								Theme.save();
							}}
						/>
					</div>
				</nav>
				<LoadingBar intent="primary"/>
				<ReactRouter.Route path="/" exact={true} render={() => (
					Constants.user ? <Vpcs/> : <Users/>
				)}/>
				<ReactRouter.Route path="/reload" render={() => (
					<ReactRouter.Redirect to="/"/>
				)}/>
				<ReactRouter.Route path="/users" render={() => (
					<Users/>
				)}/>
				<ReactRouter.Route exact path="/user" render={() => (
					<UserDetailed/>
				)}/>
				<ReactRouter.Route path="/user/:userId" render={(props) => (
					<UserDetailed userId={props.match.params.userId}/>
				)}/>
				<ReactRouter.Route path="/nodes" render={() => (
					<Nodes/>
				)}/>
				<ReactRouter.Route path="/policies" render={() => (
					<Policies/>
				)}/>
				<ReactRouter.Route path="/certificates" render={() => (
					<Certificates/>
				)}/>
				<ReactRouter.Route path="/organizations" render={() => (
					<Organizations/>
				)}/>
				<ReactRouter.Route path="/datacenters" render={() => (
					<Datacenters/>
				)}/>
				<ReactRouter.Route path="/zones" render={() => (
					<Zones/>
				)}/>
				<ReactRouter.Route path="/vpcs" render={() => (
					<Vpcs/>
				)}/>
				<ReactRouter.Route path="/domains" render={() => (
					<Domains/>
				)}/>
				<ReactRouter.Route path="/storages" render={() => (
					<Storages/>
				)}/>
				<ReactRouter.Route path="/images" render={() => (
					<Images/>
				)}/>
				<ReactRouter.Route path="/disks" render={() => (
					<Disks/>
				)}/>
				<ReactRouter.Route path="/instances" render={() => (
					<Instances/>
				)}/>
				<ReactRouter.Route path="/firewalls" render={() => (
					<Firewalls/>
				)}/>
				<ReactRouter.Route path="/authorities" render={() => (
					<Authorities/>
				)}/>
				<ReactRouter.Route path="/logs" render={() => (
					<Logs/>
				)}/>
				<ReactRouter.Route path="/settings" render={() => (
					<Settings/>
				)}/>
				<ReactRouter.Route path="/subscription" render={() => (
					<Subscription/>
				)}/>
			</div>
		</ReactRouter.HashRouter>;
	}
}
