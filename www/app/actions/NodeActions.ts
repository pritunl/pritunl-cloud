/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as NodeTypes from '../types/NodeTypes';
import * as AdvisoryTypes from '../types/AdvisoryTypes';
import NodesStore from '../stores/NodesStore';
import CompletionStore from "../stores/CompletionStore";
import * as MiscUtils from '../utils/MiscUtils';

let syncId: string;
let syncZonesId: string;
let dataSyncReqs: {[key: string]: SuperAgent.Request} = {};
let chartReqs: {[key: string]: Promise<any>} = {};

export function sync(noLoading?: boolean): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader: Loader;
	if (!noLoading) {
		loader = new Loader().loading();
	}

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/node')
			.query({
				...NodesStore.filter,
				page: NodesStore.page,
				page_count: NodesStore.pageCount,
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
					Alert.errorRes(res, 'Failed to load nodes');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: NodeTypes.SYNC,
					data: {
						nodes: res.body.nodes,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function syncZone(zone: string): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncZonesId = curSyncId;

	if (!zone) {
		Dispatcher.dispatch({
			type: NodeTypes.SYNC_ZONE,
			data: {
				nodes: [],
			},
		});
		return Promise.resolve();
	}

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/node')
			.query({
				names: true,
				zone: zone,
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

				if (curSyncId !== syncZonesId) {
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load nodes names');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: NodeTypes.SYNC_ZONE,
					data: {
						nodes: res.body,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: NodeTypes.TRAVERSE,
		data: {
			page: page,
		},
	});

	return sync();
}

export function filter(filt: NodeTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: NodeTypes.FILTER,
		data: {
			filter: filt,
		},
	});

	return sync();
}

export function commit(node: NodeTypes.Node): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/node/' + node.id)
			.send(node)
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
					Alert.errorRes(res, 'Failed to save node');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function operation(nodeId: string, operation: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/node/' + nodeId + '/' + operation)
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
					Alert.errorRes(res, 'Failed to update node');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function init(nodeId: string,
		data: NodeTypes.NodeInit): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/node/' + nodeId + '/init')
			.send(data)
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
					Alert.errorRes(res, 'Failed to update node');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(node: NodeTypes.Node): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/node')
			.send(node)
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
					Alert.errorRes(res, 'Failed to create node');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(nodeId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/node/' + nodeId)
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
					Alert.errorRes(res, 'Failed to delete nodes');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function loadAdvisories(
		nodeId: string): Promise<AdvisoryTypes.Advisory[]> {

	let loader = new Loader().loading();

	return new Promise<AdvisoryTypes.Advisory[]>((resolve, reject): void => {
		SuperAgent
			.get('/node/' + nodeId + '/advisory')
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', CompletionStore.userOrganization)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve([]);
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load node advisories');
					reject(err);
					return;
				}

				resolve(res.body || []);
			});
	});
}

export function chart(nodeId: string, resource: string,
		period: number, interval: number): Promise<any> {
	resource = resource.replace(/[0-9]/g, '');

	let cacheKey = nodeId + ':' + resource + ':' +
		period + ':' + interval;
	let existing = chartReqs[cacheKey];
	if (existing) {
		return existing;
	}

	let curDataSyncId = MiscUtils.uuid();

	let loader = new Loader().loading();

	let promise = new Promise<any>((resolve, reject): void => {
		let req = SuperAgent.get('/node/' + nodeId + '/chart')
			.query({
				resource: resource,
				period: period.toString(),
				interval: interval.toString(),
			})
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.set('Organization', CompletionStore.userOrganization)
			.on('abort', () => {
				loader.done();
				resolve(null);
			});
		dataSyncReqs[curDataSyncId] = req;

		req.end((err: any, res: SuperAgent.Response): void => {
			delete dataSyncReqs[curDataSyncId];
			loader.done();

			if (res && res.status === 401) {
				window.location.href = '/login';
				resolve(null);
				return;
			}

			if (err) {
				Alert.errorRes(res, 'Failed to load node chart');
				reject(err);
				return;
			}

			resolve(res.body);
		});
	});

	chartReqs[cacheKey] = promise;
	let clear = (): void => {
		delete chartReqs[cacheKey];
	};
	promise.then(clear, clear);

	return promise;
}

export function dataCancel(): void {
	for (let val of Object.values(dataSyncReqs)) {
		val.abort();
	}
}

EventDispatcher.register((action: NodeTypes.NodeDispatch) => {
	switch (action.type) {
		case NodeTypes.CHANGE:
			sync();
			break;
	}
});
