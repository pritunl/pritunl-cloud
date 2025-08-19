/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as DiskTypes from '../types/DiskTypes';
import DisksStore from '../stores/DisksStore';
import CompletionStore from "../stores/CompletionStore";
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
			.get('/disk')
			.query({
				...DisksStore.filter,
				page: DisksStore.page,
				page_count: DisksStore.pageCount,
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
					Alert.errorRes(res, 'Failed to load disks');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: DiskTypes.SYNC,
					data: {
						disks: res.body.disks,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: DiskTypes.TRAVERSE,
		data: {
			page: page,
		},
	});

	return sync();
}

export function filter(filt: DiskTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: DiskTypes.FILTER,
		data: {
			filter: filt,
		},
	});

	return sync();
}

export function commit(disk: DiskTypes.Disk): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/disk/' + disk.id)
			.send(disk)
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
					Alert.errorRes(res, 'Failed to save disk');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(disk: DiskTypes.Disk): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/disk')
			.send(disk)
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
					Alert.errorRes(res, 'Failed to create disk');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(diskId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/disk/' + diskId)
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
					Alert.errorRes(res, 'Failed to delete disk');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(diskIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/disk')
			.send(diskIds)
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
					Alert.errorRes(res, 'Failed to delete disks');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function forceRemoveMulti(diskIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/disk')
			.query({
				force: true,
			})
			.send(diskIds)
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
					Alert.errorRes(res, 'Failed to delete disks');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function updateMulti(diskIds: string[],
		state: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/disk')
			.send({
				"ids": diskIds,
				"state": state,
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

				if (err) {
					Alert.errorRes(res, 'Failed to update disks');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

EventDispatcher.register((action: DiskTypes.DiskDispatch) => {
	switch (action.type) {
		case DiskTypes.CHANGE:
			sync();
			break;
	}
});
