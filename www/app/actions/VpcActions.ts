/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as VpcTypes from '../types/VpcTypes';
import VpcsStore from '../stores/VpcsStore';
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
			.get('/vpc')
			.query({
				...VpcsStore.filter,
				page: VpcsStore.page,
				page_count: VpcsStore.pageCount,
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
					Alert.errorRes(res, 'Failed to load vpcs');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: VpcTypes.SYNC,
					data: {
						vpcs: res.body.vpcs,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function syncNames(): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncNamesId = curSyncId;

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/vpc')
			.query({
				names: "true",
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
					Alert.errorRes(res, 'Failed to load vpcs names');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: VpcTypes.SYNC_NAMES,
					data: {
						vpcs: res.body,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: VpcTypes.TRAVERSE,
		data: {
			page: page,
		},
	});

	return sync();
}

export function filter(filt: VpcTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: VpcTypes.FILTER,
		data: {
			filter: filt,
		},
	});

	return sync();
}

export function commit(vpc: VpcTypes.Vpc): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/vpc/' + vpc.id)
			.send(vpc)
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
					Alert.errorRes(res, 'Failed to save vpc');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(vpc: VpcTypes.Vpc): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/vpc')
			.send(vpc)
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
					Alert.errorRes(res, 'Failed to create vpc');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(vpcId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/vpc/' + vpcId)
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
					Alert.errorRes(res, 'Failed to delete vpc');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(vpcIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/vpc')
			.send(vpcIds)
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
					Alert.errorRes(res, 'Failed to delete vpcs');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

EventDispatcher.register((action: VpcTypes.VpcDispatch) => {
	switch (action.type) {
		case VpcTypes.CHANGE:
			sync();
			break;
	}
});
