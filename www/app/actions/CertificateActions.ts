/// <reference path="../References.d.ts"/>
import * as SuperAgent from 'superagent';
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Csrf from '../Csrf';
import Loader from '../Loader';
import * as CertificateTypes from '../types/CertificateTypes';
import CertificatesStore from '../stores/CertificatesStore';
import * as MiscUtils from '../utils/MiscUtils';
import * as Constants from "../Constants";
import OrganizationsStore from "../stores/OrganizationsStore";

let syncId: string;

export function sync(): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.get('/certificate')
			.query({
				...CertificatesStore.filter,
				page: CertificatesStore.page,
				page_count: CertificatesStore.pageCount,
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
					Alert.errorRes(res, 'Failed to load certificates');
					reject(err);
					return;
				}

				Dispatcher.dispatch({
					type: CertificateTypes.SYNC,
					data: {
						certificates: res.body,
						count: res.body.count,
					},
				});

				resolve();
			});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: CertificateTypes.TRAVERSE,
		data: {
			page: page,
		},
	});
	return sync();
}

export function filter(filt: CertificateTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: CertificateTypes.FILTER,
		data: {
			filter: filt,
		},
	});
	return sync();
}

export function commit(cert: CertificateTypes.Certificate): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.put('/certificate/' + cert.id)
			.send(cert)
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
					Alert.errorRes(res, 'Failed to save certificate');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function create(cert: CertificateTypes.Certificate): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.post('/certificate')
			.send(cert)
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
					Alert.errorRes(res, 'Failed to create certificate');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function remove(certId: string): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/certificate/' + certId)
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
					Alert.errorRes(res, 'Failed to delete certificates');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

export function removeMulti(certificateIds: string[]): Promise<void> {
	let loader = new Loader().loading();

	return new Promise<void>((resolve, reject): void => {
		SuperAgent
			.delete('/certificate')
			.send(certificateIds)
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
					Alert.errorRes(res, 'Failed to delete certificates');
					reject(err);
					return;
				}

				resolve();
			});
	});
}

EventDispatcher.register((action: CertificateTypes.CertificateDispatch) => {
	switch (action.type) {
		case CertificateTypes.CHANGE:
			sync();
			break;
	}
});
