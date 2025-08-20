/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as StorageTypes from '../types/StorageTypes';
import StoragesStore from '../stores/StoragesStore';
import * as MiscUtils from '../utils/MiscUtils';
import * as Constants from "../Constants";

let syncId: string;

export function sync(): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/storage')
			.query({
				...StoragesStore.filter,
				page: StoragesStore.page,
				page_count: StoragesStore.pageCount,
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

				if (curSyncId !== syncId) {
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load storages');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: StorageTypes.SYNC,
					data: {
						storages: res.body.storages,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: StorageTypes.TRAVERSE,
		data: {
			page: page,
		},
	});
	return sync();
}

export function filter(filt: StorageTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: StorageTypes.FILTER,
		data: {
			filter: filt,
		},
	});
	return sync();
}

export function commit(storage: StorageTypes.Storage): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/storage/' + storage.id)
			.send(storage)
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
					Alert.errorRes(res, 'Failed to save storage');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(storage: StorageTypes.Storage): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/storage')
			.send(storage)
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
					Alert.errorRes(res, 'Failed to create storage');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(storageId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/storage/' + storageId)
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
					Alert.errorRes(res, 'Failed to delete storages');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(storageIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/storage')
			.send(storageIds)
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
					Alert.errorRes(res, 'Failed to delete storages');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

EventDispatcher.register((action: StorageTypes.StorageDispatch) => {
	switch (action.type) {
		case StorageTypes.CHANGE:
			if (!Constants.user) {
				sync();
			}
			break;
	}
});
