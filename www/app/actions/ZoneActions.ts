/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as ZoneTypes from '../types/ZoneTypes';
import ZonesStore from '../stores/ZonesStore';
import CompletionStore from "../stores/CompletionStore";
import * as MiscUtils from '../utils/MiscUtils';
import * as Constants from "../Constants";

let syncId: string;
let syncNamesId: string;

export function sync(): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/zone')
			.query({
				...ZonesStore.filter,
				page: ZonesStore.page,
				page_count: ZonesStore.pageCount,
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

				if (curSyncId !== syncId) {
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load zones');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: ZoneTypes.SYNC,
					data: {
						zones: res.body.zones,
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
			.get('/zone')
			.query({
				names: true,
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

				if (curSyncId !== syncNamesId) {
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to load zone names');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: ZoneTypes.SYNC_NAMES,
					data: {
						secrets: res.body,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: ZoneTypes.TRAVERSE,
		data: {
			page: page,
		},
	});
	return sync();
}

export function filter(filt: ZoneTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: ZoneTypes.FILTER,
		data: {
			filter: filt,
		},
	});
	return sync();
}

export function commit(zone: ZoneTypes.Zone): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/zone/' + zone.id)
			.send(zone)
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
					Alert.errorRes(res, 'Failed to save zone');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(zone: ZoneTypes.Zone): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/zone')
			.send(zone)
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
					Alert.errorRes(res, 'Failed to create zone');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(zoneId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/zone/' + zoneId)
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
					Alert.errorRes(res, 'Failed to delete zones');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(zoneIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/zone')
			.send(zoneIds)
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
					Alert.errorRes(res, 'Failed to delete zones');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

EventDispatcher.register((action: ZoneTypes.ZoneDispatch) => {
	switch (action.type) {
		case ZoneTypes.CHANGE:
			if (!Constants.user) {
				sync();
			}
			break;
	}
});
