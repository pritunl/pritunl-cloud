/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as AdvisoryTypes from '../types/AdvisoryTypes';
import AdvisoriesStore from '../stores/AdvisoriesStore';
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
			.get('/advisory')
			.query({
				...AdvisoriesStore.filter,
				page: AdvisoriesStore.page,
				page_count: AdvisoriesStore.pageCount,
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
					Alert.errorRes(res, 'Failed to load advisories');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: AdvisoryTypes.SYNC,
					data: {
						advisories: res.body.advisories,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: AdvisoryTypes.TRAVERSE,
		data: {
			page: page,
		},
	});

	return sync();
}

export function filter(filt: AdvisoryTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: AdvisoryTypes.FILTER,
		data: {
			filter: filt,
		},
	});

	return sync();
}

function dismissUpdate(advisoryId: string,
		data: AdvisoryTypes.DismissData): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/advisory/' + advisoryId + '/dismiss')
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
					Alert.errorRes(res, 'Failed to update advisory');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function dismiss(advisoryId: string): Promise<void> {
	return dismissUpdate(advisoryId, {
		dismiss: true,
	});
}

export function restore(advisoryId: string): Promise<void> {
	return dismissUpdate(advisoryId, {
		restore: true,
	});
}

export function dismissResources(advisoryId: string,
		dismissals: string[]): Promise<void> {
	return dismissUpdate(advisoryId, {
		dismissals: dismissals,
	});
}

export function restoreResources(advisoryId: string,
		restores: string[]): Promise<void> {
	return dismissUpdate(advisoryId, {
		restores: restores,
	});
}

export function remove(advisoryId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/advisory/' + advisoryId)
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
					Alert.errorRes(res, 'Failed to delete advisory');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(advisoryIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/advisory')
			.send(advisoryIds)
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
					Alert.errorRes(res, 'Failed to delete advisories');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

EventDispatcher.register((action: AdvisoryTypes.AdvisoryDispatch) => {
	switch (action.type) {
		case AdvisoryTypes.CHANGE:
			sync();
			break;
	}
});
