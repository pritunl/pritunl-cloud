/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as ImageTypes from '../types/ImageTypes';
import ImagesStore from '../stores/ImagesStore';
import CompletionStore from "../stores/CompletionStore";
import * as MiscUtils from '../utils/MiscUtils';

let syncId: string;

export function sync(): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/image')
			.query({
				...ImagesStore.filter,
				page: ImagesStore.page,
				page_count: ImagesStore.pageCount,
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
					Alert.errorRes(res, 'Failed to load images');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: ImageTypes.SYNC,
					data: {
						images: res.body.images,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function syncDatacenter(datacenter: string): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	if (!datacenter) {
		Dispatcher.dispatch({
			type: ImageTypes.SYNC_DATACENTER,
			data: {
				images: [],
			},
		});
		return Promise.resolve();
	}

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/image')
			.query({
				datacenter: datacenter,
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
					Alert.errorRes(res, 'Failed to load images names');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: ImageTypes.SYNC_DATACENTER,
					data: {
						images: res.body,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: ImageTypes.TRAVERSE,
		data: {
			page: page,
		},
	});

	return sync();
}

export function filter(filt: ImageTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: ImageTypes.FILTER,
		data: {
			filter: filt,
		},
	});

	return sync();
}

export function commit(image: ImageTypes.Image): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/image/' + image.id)
			.send(image)
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
					Alert.errorRes(res, 'Failed to save image');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(image: ImageTypes.Image): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/image')
			.send(image)
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
					Alert.errorRes(res, 'Failed to create image');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(imageId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/image/' + imageId)
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
					Alert.errorRes(res, 'Failed to delete image');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(imageIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/image')
			.send(imageIds)
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
					Alert.errorRes(res, 'Failed to delete images');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

EventDispatcher.register((action: ImageTypes.ImageDispatch) => {
	switch (action.type) {
		case ImageTypes.CHANGE:
			sync();
			break;
	}
});
