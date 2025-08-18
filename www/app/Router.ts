/// <reference path="./References.d.ts"/>
import * as CompletionActions from './actions/CompletionActions';
import * as UserActions from './actions/UserActions';
import * as SessionActions from './actions/SessionActions';
import * as AuditActions from './actions/AuditActions';
import * as NodeActions from './actions/NodeActions';
import * as PolicyActions from './actions/PolicyActions';
import * as CertificateActions from './actions/CertificateActions';
import * as SecretActions from './actions/SecretActions';
import * as OrganizationActions from './actions/OrganizationActions';
import * as DatacenterActions from './actions/DatacenterActions';
import * as AlertActions from './actions/AlertActions';
import * as ZoneActions from './actions/ZoneActions';
import * as ShapeActions from './actions/ShapeActions';
import * as BlockActions from './actions/BlockActions';
import * as VpcActions from './actions/VpcActions';
import * as DomainActions from './actions/DomainActions';
import * as PlanActions from './actions/PlanActions';
import * as BalancerActions from './actions/BalancerActions';
import * as StorageActions from './actions/StorageActions';
import * as ImageActions from './actions/ImageActions';
import * as PoolActions from './actions/PoolActions';
import * as DiskActions from './actions/DiskActions';
import * as InstanceActions from './actions/InstanceActions';
import * as PodActions from './actions/PodActions';
import * as FirewallActions from './actions/FirewallActions';
import * as AuthorityActions from './actions/AuthorityActions';
import * as LogActions from './actions/LogActions';
import * as SettingsActions from './actions/SettingsActions';
import * as SubscriptionActions from './actions/SubscriptionActions';

export function setLocation(location: string) {
	window.location.hash = location
	let evt = new Event("router_update")
	window.dispatchEvent(evt)
}

export function reload() {
	let evt = new Event("router_update")
	window.dispatchEvent(evt)
}

export function refresh(callback?: () => void) {
	let pathname = window.location.hash.replace(/^#/, '');
	CompletionActions.sync();

	if (pathname === '/users') {
		UserActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname.startsWith('/user/')) {
		UserActions.reload().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
		SessionActions.reload().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
		AuditActions.reload().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/nodes') {
		NodeActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/policies') {
		SettingsActions.sync();
		PolicyActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/certificates') {
		CertificateActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/secrets') {
		SecretActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/organizations') {
		OrganizationActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/datacenters') {
		DatacenterActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/zones') {
		ZoneActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/shapes') {
		ShapeActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/blocks') {
		BlockActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/vpcs') {
		VpcActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/domains') {
		DomainActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/plans') {
		PlanActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/balancers') {
		BalancerActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/storages') {
		StorageActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/images') {
		ImageActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/pools') {
		PoolActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/disks') {
		DiskActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/instances') {
		InstanceActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/pods') {
		PodActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/firewalls') {
		FirewallActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/authorities') {
		AuthorityActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/alerts') {
		AlertActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/logs') {
		LogActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/settings') {
		SettingsActions.sync().then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else if (pathname === '/subscription') {
		SubscriptionActions.sync(true).then((): void => {
			if (callback) {
				callback()
			}
		}).catch((): void => {
			if (callback) {
				callback()
			}
		});
	} else {
		console.log(`Failed to match refresh ${pathname}`)
		this.setState({
			...this.state,
			disabled: false,
		});
	}
}
