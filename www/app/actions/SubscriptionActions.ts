/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as SubscriptionTypes from '../types/SubscriptionTypes';
import * as MiscUtils from '../utils/MiscUtils';
import * as Constants from "../Constants";

let syncId: string;

export function sync(update: boolean): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/subscription' + (update ? '/update' : ''))
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
					Alert.errorRes(res, 'Failed to sync subscription');
					reject(err);

					Dispatcher.dispatch({
						type: SubscriptionTypes.SYNC,
						data: {},
					});

					return;
				}

				Dispatcher.dispatch({
					type: SubscriptionTypes.SYNC,
					data: res.body,
				});

				resolve();
			});
	});
}

export function activate(license: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/subscription')
			.send({
				license: license,
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
					Alert.errorRes(res, 'Failed to activate subscription');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: SubscriptionTypes.SYNC,
					data: res.body,
				});

				resolve();
			});
	});
}

export function cancel(key: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('https://app.pritunl.com/subscription')
			.send({
				key: key,
			})
			.set('Accept', 'application/json')
			.end((err: any, res: SuperAgent.Response): void => {
				loader.done();

				if (res && res.status === 401) {
					window.location.href = '/login';
					resolve();
					return;
				}

				if (err) {
					Alert.errorRes(res, 'Failed to cancel subscription');
					reject(err);
					return;
				}

				resolve();

				sync(true);
			});
	});
}

EventDispatcher.register((action: SubscriptionTypes.SubscriptionDispatch) => {
	switch (action.type) {
		case SubscriptionTypes.CHANGE:
			if (!Constants.user) {
				sync(false);
			}
			break;
	}
});
