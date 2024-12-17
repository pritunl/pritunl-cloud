/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as ServiceTypes from '../types/ServiceTypes';
import ServicesStore from '../stores/ServicesStore';
import * as MiscUtils from '../utils/MiscUtils';

let syncId: string;
let syncUnitId: string;
let lastServiceId: string;
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
			.get('/service')
			.query({
				...ServicesStore.filter,
				page: ServicesStore.page,
				page_count: ServicesStore.pageCount,
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
					Alert.errorRes(res, 'Failed to load services');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: ServiceTypes.SYNC,
					data: {
						services: res.body.services,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: ServiceTypes.TRAVERSE,
		data: {
			page: page,
		},
	});

	return sync();
}

export function filter(filt: ServiceTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: ServiceTypes.FILTER,
		data: {
			filter: filt,
		},
	});

	return sync();
}

export function commit(service: ServiceTypes.Service): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/service/' + service.id)
			.send(service)
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
					Alert.errorRes(res, 'Failed to save service');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(service: ServiceTypes.Service): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/service')
			.send(service)
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
					Alert.errorRes(res, 'Failed to create service');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(serviceId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/service/' + serviceId)
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
					Alert.errorRes(res, 'Failed to delete service');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(serviceIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/service')
			.send(serviceIds)
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
					Alert.errorRes(res, 'Failed to delete services');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function syncUnit(serviceId?: string, unitId?: string): Promise<void> {
	if (!serviceId) {
		serviceId = lastServiceId
	} else {
		lastServiceId = serviceId
	}

	if (!unitId) {
		unitId = lastUnitId
	} else {
		lastUnitId = unitId
	}

	if (!serviceId || !unitId) {
		return Promise.resolve();
	}

	let curSyncId = MiscUtils.uuid();
	syncUnitId = curSyncId;

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/service/' + serviceId + "/unit/" + unitId)
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
					Alert.errorRes(res, 'Failed to load service unit');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: ServiceTypes.SYNC_UNIT,
					data: {
						unit: res.body,
					},
				});

				resolve();
			});
	});
}

export function deployUnit(serviceId: string, unitId: string,
	specId: string, count: number): Promise<void> {

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/service/' + serviceId + "/unit/" + unitId + "/deployment")
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

export function updateMultiUnitState(serviceId: string, unitId: string,
	deploymentIds: string[], state: string): Promise<void> {

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/service/' + serviceId + "/unit/" + unitId + "/deployment")
			.query({
				state: state,
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

export function log(deply: ServiceTypes.Deployment,
	resource: string): Promise<any> {

	let curDataSyncId = MiscUtils.uuid();

	let loader = new Loader().loading();

	return new Promise<any>((resolve, reject): void => {
		let req = SuperAgent.get('/service/' + deply.service +
				"/unit/" + deply.unit + "/deployment/" + deply.id + "/log")
			.query({
				resource: resource,
			})
			.set('Accept', 'application/json')
			.set('Csrf-Token', Csrf.token)
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
				Alert.errorRes(res, 'Failed to load check log');
				reject(err);
				return;
			}

			resolve(res.body);
		});
	});
}

export function dataCancel(): void {
	for (let [key, val] of Object.entries(dataSyncReqs)) {
		val.abort();
	}
}

EventDispatcher.register((action: ServiceTypes.ServiceDispatch) => {
	switch (action.type) {
		case ServiceTypes.CHANGE:
			sync();
			syncUnit();
			break;
	}
});
