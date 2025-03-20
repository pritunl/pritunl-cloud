/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as PodTypes from '../types/PodTypes';
import PodsStore from '../stores/PodsStore';
import * as MiscUtils from '../utils/MiscUtils';

let syncId: string;
let syncUnitId: string;
let lastPodId: string;
let lastUnitId: string
let dataSyncReqs: {[key: string]: SuperAgent.Request} = {};

export function sync(noLoading?: boolean): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader: Loader;
	if (!noLoading) {
		loader = new Loader().loading();
	}

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/pod')
			.query({
				...PodsStore.filter,
				page: PodsStore.page,
				page_count: PodsStore.pageCount,
			})
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
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
					Alert.errorRes(res, 'Failed to load pods');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: PodTypes.SYNC,
					data: {
						pods: res.body.pods,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: PodTypes.TRAVERSE,
		data: {
			page: page,
		},
	});

	return sync();
}

export function filter(filt: PodTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: PodTypes.FILTER,
		data: {
			filter: filt,
		},
	});

	return sync();
}

export function commit(pod: PodTypes.Pod): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/pod/' + pod.id)
			.send(pod)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to save pod');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(pod: PodTypes.Pod): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/pod')
			.send(pod)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to create pod');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(podId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/pod/' + podId)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to delete pod');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(podIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/pod')
			.send(podIds)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to delete pods');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function syncUnit(podId?: string, unitId?: string): Promise<void> {
	if (!podId) {
		podId = lastPodId
	} else {
		lastPodId = podId
	}

	if (!unitId) {
		unitId = lastUnitId
	} else {
		lastUnitId = unitId
	}

	if (!podId || !unitId) {
		return Promise.resolve();
	}

	let curSyncId = MiscUtils.uuid();
	syncUnitId = curSyncId;

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/pod/' + podId + "/unit/" + unitId)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (curSyncId !== syncUnitId) {
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load pod unit');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: PodTypes.SYNC_UNIT,
					data: {
						unit: res.body,
					},
				});

				resolve();
			});
	});
}

export function deployUnit(podId: string, unitId: string,
	specId: string, count: number): Promise<void> {

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/pod/' + podId + "/unit/" + unitId + "/deployment")
			.send({
				count: count,
				spec: specId,
			})
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to create deployments');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function updateMultiUnitAction(podId: string, unitId: string,
	deploymentIds: string[], action: string, commit?: string): Promise<void> {

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/pod/' + podId + "/unit/" + unitId + "/deployment")
			.query({
				action: action,
				commit: commit,
			})
			.send(deploymentIds)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to delete deployments');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function commitDeployment(deply: PodTypes.Deployment): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/pod/' + deply.pod + "/unit/" + deply.unit +
				"/deployment/" + deply.id)
			.send(deply)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to save deployment');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function log(deply: PodTypes.Deployment,
	resource: string, noLoading?: boolean): Promise<any> {

	let curDataSyncId = MiscUtils.uuid();

	let loader: Loader;
	if (!noLoading) {
		loader = new Loader().loading();
	}

	return new Promise<any>((resolve, reject): void => {
		let req = SuperAgent.get('/pod/' + deply.pod +
				"/unit/" + deply.unit + "/deployment/" + deply.id + "/log")
			.query({
				resource: resource,
			})
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.on('abort', () => {
				if (loader) {
					loader.done();
				}
				resolve(null);
			});
		dataSyncReqs[curDataSyncId] = req;

		req.end((err: any, res: SuperAgent.Response): void => {
			delete dataSyncReqs[curDataSyncId];
			if (loader) {
				loader.done();
			}

			if (res && res.status === 401) {
				window.location.href = '/login';
				resolve(null);
				return;
			}

			if (err) {
				Alert.errorRes(res, 'Failed to load check log');
				reject(err);
				return;
			}

			resolve(res.body);
		});
	});
}

export function syncSpecs(podId: string, unitId: string, page: number,
	pageCount: number, noLoading?: boolean): Promise<PodTypes.CommitData> {

	let loader: Loader;
	if (!noLoading) {
		loader = new Loader().loading();
	}

	return new Promise<PodTypes.CommitData>((resolve, reject): void => {
		SuperAgent
			.get("/pod/" + podId + "/unit/" + unitId + "/spec")
			.query({
				page: page,
				page_count: pageCount,
			})
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				if (loader) {
					loader.done();
				}

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve(null);
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load unit commits');
					reject(err);
					return;
				}

				res.body.unit = unitId
				res.body.page = page
				res.body.page_count = pageCount
				resolve(res.body as PodTypes.CommitData);
			});
	});
}

export function spec(podId: string, unitId: string,
	specId: string): Promise<PodTypes.Commit> {

	return new Promise<PodTypes.Commit>((resolve, reject): void => {
		SuperAgent
			.get("/pod/" + podId + "/unit/" + unitId + "/spec/" + specId)
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
			.end((err: any, res: SuperAgent.Response): void => {
				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve(null);
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load unit commits');
					reject(err);
					return;
				}

				resolve(res.body as PodTypes.Commit);
			});
	});
}

export function dataCancel(): void {
	for (let [key, val] of Object.entries(dataSyncReqs)) {
		val.abort();
	}
}

EventDispatcher.register((action: PodTypes.PodDispatch) => {
	switch (action.type) {
		case PodTypes.CHANGE:
			sync();
			syncUnit();
			break;
	}
});
