/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as SecretTypes from '../types/SecretTypes';
import SecretsStore from '../stores/SecretsStore';
import * as MiscUtils from '../utils/MiscUtils';
import * as Constants from "../Constants";
import OrganizationsStore from "../stores/OrganizationsStore";

let syncId: string;
let syncNamesId: string;

export function sync(): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/secret')
			.query({
				...SecretsStore.filter,
				page: SecretsStore.page,
				page_count: SecretsStore.pageCount,
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
					Alert.errorRes(res, 'Failed to load secrets');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: SecretTypes.SYNC,
					data: {
						secrets: res.body,
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
			.get('/secret')
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
					Alert.errorRes(res, 'Failed to load secret names');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: SecretTypes.SYNC_NAMES,
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
		type: SecretTypes.TRAVERSE,
		data: {
			page: page,
		},
	});
	return sync();
}

export function filter(filt: SecretTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: SecretTypes.FILTER,
		data: {
			filter: filt,
		},
	});
	return sync();
}

export function commit(secr: SecretTypes.Secret): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/secret/' + secr.id)
			.send(secr)
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
					Alert.errorRes(res, 'Failed to save secret');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(secr: SecretTypes.Secret): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/secret')
			.send(secr)
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
					Alert.errorRes(res, 'Failed to create secret');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(secrId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/secret/' + secrId)
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
					Alert.errorRes(res, 'Failed to delete secrets');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(secretIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/secret')
			.send(secretIds)
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
					Alert.errorRes(res, 'Failed to delete secrets');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

EventDispatcher.register((action: SecretTypes.SecretDispatch) => {
	switch (action.type) {
		case SecretTypes.CHANGE:
			sync();
			break;
	}
});
