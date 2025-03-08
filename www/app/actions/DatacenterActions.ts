/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as DatacenterTypes from '../types/DatacenterTypes';
import DatacentersStore from '../stores/DatacentersStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import * as MiscUtils from '../utils/MiscUtils';
import * as Constants from "../Constants";

let syncId: string;

export function sync(): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/datacenter')
			.query({
				...DatacentersStore.filter,
				page: DatacentersStore.page,
				page_count: DatacentersStore.pageCount,
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

				if (curSyncId !== syncId) {
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load datacenters');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: DatacenterTypes.SYNC,
					data: {
						datacenters: res.body,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: DatacenterTypes.TRAVERSE,
		data: {
			page: page,
		},
	});
	return sync();
}

export function filter(filt: DatacenterTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: DatacenterTypes.FILTER,
		data: {
			filter: filt,
		},
	});
	return sync();
}

export function commit(datacenter: DatacenterTypes.Datacenter): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/datacenter/' + datacenter.id)
			.send(datacenter)
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
					Alert.errorRes(res, 'Failed to save datacenter');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(datacenter: DatacenterTypes.Datacenter): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/datacenter')
			.send(datacenter)
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
					Alert.errorRes(res, 'Failed to create datacenter');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(datacenterId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/datacenter/' + datacenterId)
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
					Alert.errorRes(res, 'Failed to delete datacenters');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(datacenterIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/datacenter')
			.send(datacenterIds)
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
					Alert.errorRes(res, 'Failed to delete datacenters');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

EventDispatcher.register((action: DatacenterTypes.DatacenterDispatch) => {
	switch (action.type) {
		case DatacenterTypes.CHANGE:
			if (!Constants.user) {
				sync();
			}
			break;
	}
});
