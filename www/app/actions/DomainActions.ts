/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as DomainTypes from '../types/DomainTypes';
import DomainsStore from '../stores/DomainsStore';
import CompletionStore from "../stores/CompletionStore";
import * as MiscUtils from '../utils/MiscUtils';

let syncId: string;
let syncNamesId: string;

export function sync(noLoading?: boolean): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader: Loader;
	if (!noLoading) {
		loader = new Loader().loading();
	}

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/domain')
			.query({
				...DomainsStore.filter,
				page: DomainsStore.page,
				page_count: DomainsStore.pageCount,
			})
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', CompletionStore.userOrganization)
			.end((err: any, res: SuperAgent.Response): void => {
				if (loader) {
					loader.done();
				}

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (curSyncId !== syncId) {
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load domains');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: DomainTypes.SYNC,
					data: {
						domains: res.body.domains,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function syncName(): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncNamesId = curSyncId;

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/domain')
			.query({
				names: true,
			})
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', CompletionStore.userOrganization)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (curSyncId !== syncNamesId) {
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load domain names');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: DomainTypes.SYNC_NAME,
					data: {
						domains: res.body,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: DomainTypes.TRAVERSE,
		data: {
			page: page,
		},
	});

	return sync();
}

export function filter(filt: DomainTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: DomainTypes.FILTER,
		data: {
			filter: filt,
		},
	});

	return sync();
}

export function commit(domain: DomainTypes.Domain): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/domain/' + domain.id)
			.send(domain)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', CompletionStore.userOrganization)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to save domain');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(domain: DomainTypes.Domain): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/domain')
			.send(domain)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', CompletionStore.userOrganization)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to create domain');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(domainId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/domain/' + domainId)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', CompletionStore.userOrganization)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to delete domain');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(domainIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/domain')
			.send(domainIds)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', CompletionStore.userOrganization)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to delete domains');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

EventDispatcher.register((action: DomainTypes.DomainDispatch) => {
	switch (action.type) {
		case DomainTypes.CHANGE:
			sync();
			break;
	}
});
