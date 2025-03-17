/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as InstanceTypes from '../types/InstanceTypes';
import InstancesStore from '../stores/InstancesStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import * as MiscUtils from '../utils/MiscUtils';

let syncId: string;

export function sync(noLoading?: boolean): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader: Loader;
	if (!noLoading) {
		loader = new Loader().loading();
	}

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/instance')
			.query({
				...InstancesStore.filter,
				page: InstancesStore.page,
				page_count: InstancesStore.pageCount,
			})
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', OrganizationsStore.current)
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
					Alert.errorRes(res, 'Failed to load instances');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: InstanceTypes.SYNC,
					data: {
						instances: res.body.instances,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: InstanceTypes.TRAVERSE,
		data: {
			page: page,
		},
	});

	return sync();
}

export function filter(filt: InstanceTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: InstanceTypes.FILTER,
		data: {
			filter: filt,
		},
	});

	return sync();
}

export function commit(instance: InstanceTypes.Instance): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/instance/' + instance.id)
			.send(instance)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', OrganizationsStore.current)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to save instance');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(instance: InstanceTypes.Instance): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/instance')
			.send(instance)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', OrganizationsStore.current)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to create instance');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(instanceId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/instance/' + instanceId)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', OrganizationsStore.current)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to delete instance');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(instanceIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/instance')
			.send(instanceIds)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', OrganizationsStore.current)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to delete instances');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function forceRemoveMulti(instanceIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/instance')
			.query({
				force: true,
			})
			.send(instanceIds)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', OrganizationsStore.current)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to force delete instances');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function updateMulti(instanceIds: string[],
		action: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/instance')
			.send({
				"ids": instanceIds,
				"action": action,
			})
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', OrganizationsStore.current)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to update instances');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function syncNode(node: string, pool: string): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let scope: string;
	let query: {[key: string]: string};
	if (node) {
		scope = node;
		query = {
			node_names: node,
		};
	} else {
		scope = pool;
		query = {
			pool_names: pool,
		};
	}

	if (!scope) {
		return Promise.resolve();
	}

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/instance')
			.query(query)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', OrganizationsStore.current)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

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
					Alert.errorRes(res, 'Failed to load instance names');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: InstanceTypes.SYNC_NODE,
					data: {
						scope: scope,
						instances: res.body,
					},
				});

				resolve();
			});
	});
}

EventDispatcher.register((action: InstanceTypes.InstanceDispatch) => {
	switch (action.type) {
		case InstanceTypes.CHANGE:
			sync();
			break;
	}
});
