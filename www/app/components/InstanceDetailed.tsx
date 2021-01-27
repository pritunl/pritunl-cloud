/// <reference path="../References.d.ts"/>
import * as React from 'react';
import RFB from 'novnc-core';
import * as InstanceTypes from '../types/InstanceTypes';
import * as InstanceActions from '../actions/InstanceActions';
import * as VpcTypes from '../types/VpcTypes';
import * as DomainTypes from '../types/DomainTypes';
import * as PageInfos from './PageInfo';
import * as Csrf from '../Csrf';
import OrganizationsStore from '../stores/OrganizationsStore';
import ZonesStore from '../stores/ZonesStore';
import PageInput from './PageInput';
import PageInputButton from './PageInputButton';
import PageInfo from './PageInfo';
import PageSwitch from './PageSwitch';
import PageSelect from './PageSelect';
import PageSave from './PageSave';
import PageNumInput from './PageNumInput';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import PageSelectButton from "./PageSelectButton";
import PageTextArea from "./PageTextArea";

interface Props {
	vpcs: VpcTypes.VpcsRo;
	domains: DomainTypes.DomainsRo;
	instance: InstanceTypes.InstanceRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	instance: InstanceTypes.Instance;
	addCert: string;
	addNetworkRole: string;
	addVpc: string;
	addDriveDevice: string;
	addUsbDevice: string;
	addPciDevice: string;
	forwardedChecked: boolean;
	vnc: boolean;
	vncCtrl: boolean;
	vncAlt: boolean;
	vncSuper: boolean;
	vncScale: boolean;
	vncHeight: number;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 0 10px',
		width: '100%',
		maxWidth: '1062px',
	} as React.CSSProperties,
	button: {
		height: '30px',
	} as React.CSSProperties,
	controlButton: {
		marginRight: '10px',
	} as React.CSSProperties,
	buttons: {
		cursor: 'pointer',
		position: 'absolute',
		top: 0,
		left: 0,
		right: 0,
		padding: '4px',
		height: '39px',
		backgroundColor: 'rgba(0, 0, 0, 0.13)',
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	itemsLabel: {
		display: 'block',
	} as React.CSSProperties,
	itemsAdd: {
		margin: '8px 0 15px 0',
	} as React.CSSProperties,
	group: {
		flex: 1,
		minWidth: '280px',
		margin: '0 10px',
	} as React.CSSProperties,
	save: {
		paddingBottom: '10px',
	} as React.CSSProperties,
	label: {
		width: '100%',
		maxWidth: '280px',
	} as React.CSSProperties,
	status: {
		margin: '6px 0 0 1px',
	} as React.CSSProperties,
	icon: {
		marginRight: '3px',
	} as React.CSSProperties,
	inputGroup: {
		width: '100%',
	} as React.CSSProperties,
	protocol: {
		flex: '0 1 auto',
	} as React.CSSProperties,
	port: {
		flex: '1',
	} as React.CSSProperties,
	select: {
		margin: '7px 0px 0px 6px',
		paddingTop: '3px',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	vncBox: {
		paddingBottom: '10px',
		position: 'relative',
	} as React.CSSProperties,
};

export default class InstanceDetailed extends React.Component<Props, State> {
	vncState: boolean;
	vncRef: React.RefObject<HTMLDivElement>;
	vncRfb: RFB;

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			instance: null,
			addCert: null,
			addNetworkRole: '',
			addVpc: '',
			addDriveDevice: '',
			addUsbDevice: '',
			addPciDevice: '',
			forwardedChecked: false,
			vnc: false,
			vncCtrl: false,
			vncAlt: false,
			vncSuper: false,
			vncScale: false,
			vncHeight: null,
		};

		this.vncRef = React.createRef();
	}

	componentDidMount(): void {
		this.vncState = true;
	}

	componentWillUnmount(): void {
		this.vncState = false;
		if (this.vncRfb) {
			this.vncRfb.disconnect();
		}
	}

	set(name: string, val: any): void {
		let instance: any;

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		instance[name] = val;

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
		});
	}

	onTogleVnc = (): void => {
		if (this.state.vnc) {
			if (this.vncRfb) {
				this.vncRfb.disconnect();
			}
		} else {
			this.connectVnc();
		}

		this.setState({
			...this.state,
			vnc: !this.state.vnc,
		});
	}

	connectVnc = (): void => {
		this.vncRfb = new RFB(
			this.vncRef.current,
			'wss://' + location.hostname + (
				location.port ? ':' + location.port : '') + '/instance/' +
			this.props.instance.id + '/vnc?csrf_token=' + Csrf.token,
			{
				shared: true,
				wsProtocols: ['binary'],
				credentials: {
					password: this.props.instance.vnc_password,
				},
			},
		);
		this.vncRfb.addEventListener('disconnect', function() {
			setTimeout(function() {
				if (this.state.vnc && this.vncState) {
					this.connectVnc();
				}
			}.bind(this), 250);
		}.bind(this));
		if (this.state.vncScale) {
			this.vncRfb.scaleViewport = 'scale';
		}
	}

	onToggleVncCtrl = (): void => {
		if (this.vncRfb) {
			if (this.state.vncCtrl) {
				this.vncRfb.sendKey(0xffe3, 'ControlLeft', false);
			} else {
				this.vncRfb.sendKey(0xffe3, 'ControlLeft', true);
			}
		}

		this.setState({
			...this.state,
			vncCtrl: !this.state.vncCtrl,
		});
	}

	onToggleVncAlt = (): void => {
		if (this.vncRfb) {
			if (this.state.vncAlt) {
				this.vncRfb.sendKey(0xffe9, 'AltLeft', false);
			} else {
				this.vncRfb.sendKey(0xffe9, 'AltLeft', true);
			}
		}

		this.setState({
			...this.state,
			vncAlt: !this.state.vncAlt,
		});
	}

	onToggleVncSuper = (): void => {
		if (this.vncRfb) {
			if (this.state.vncSuper) {
				this.vncRfb.sendKey(0xffeb, 'MetaLeft', false);
			} else {
				this.vncRfb.sendKey(0xffeb, 'MetaLeft', true);
			}
		}

		this.setState({
			...this.state,
			vncSuper: !this.state.vncSuper,
		});
	}

	onVncCtrlAltDel = (): void => {
		if (this.vncRfb) {
			this.vncRfb.sendCtrlAltDel();
		}
	}

	onVncTab = (): void => {
		if (this.vncRfb) {
			this.vncRfb.sendKey(0xff09, 'Tab');
		}
	}

	onVncEsc = (): void => {
		if (this.vncRfb) {
			this.vncRfb.sendKey(0xff1b, 'Escape');
		}
	}

	onToggleVncFullscreen = (): void => {
		if (document.fullscreenElement) {
			if (document.exitFullscreen) {
				document.exitFullscreen();
			}
		} else {
			if (this.vncRef) {
				if (this.vncRef.current.requestFullscreen) {
					this.vncRef.current.requestFullscreen();
				}
			}
		}
	}

	onToggleVncScale = (): void => {
		let vncHeight: number;
		let vncScale = this.state.vncScale;

		if (vncScale) {
			this.vncRfb.scaleViewport = '';
		} else {
			let ratio = this.vncRfb._canvas.height / this.vncRfb._canvas.width;
			vncHeight = Math.floor(this.vncRef.current.offsetWidth * ratio);
		}

		this.setState({
			...this.state,
			vncScale: !this.state.vncScale,
			vncHeight: vncHeight,
		});

		if (!vncScale) {
			this.vncRfb.scaleViewport = 'scale';
			setTimeout((): void => {
				if (this.state.vncScale) {
					this.vncRfb.scaleViewport = 'scale';
				}
				setTimeout((): void => {
					if (this.state.vncScale) {
						this.vncRfb.scaleViewport = 'scale';
					}
					setTimeout((): void => {
						if (this.state.vncScale) {
							this.vncRfb.scaleViewport = 'scale';
						}
						setTimeout((): void => {
							if (this.state.vncScale) {
								this.vncRfb.scaleViewport = 'scale';
							}
							setTimeout((): void => {
								if (this.state.vncScale) {
									this.vncRfb.scaleViewport = 'scale';
								}
								setTimeout((): void => {
									if (this.state.vncScale) {
										this.vncRfb.scaleViewport = 'scale';
									}
								}, 50);
							}, 50);
						}, 50);
					}, 50);
				}, 50);
			}, 50);
		}
	}

	onAddNetworkRole = (): void => {
		let instance: InstanceTypes.Instance;

		if (!this.state.addNetworkRole) {
			return;
		}

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		let networkRoles = [
			...(instance.network_roles || []),
		];

		if (networkRoles.indexOf(this.state.addNetworkRole) === -1) {
			networkRoles.push(this.state.addNetworkRole);
		}

		networkRoles.sort();
		instance.network_roles = networkRoles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addNetworkRole: '',
			instance: instance,
		});
	}

	onRemoveNetworkRole = (networkRole: string): void => {
		let instance: InstanceTypes.Instance;

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		let networkRoles = [
			...(instance.network_roles || []),
		];

		let i = networkRoles.indexOf(networkRole);
		if (i === -1) {
			return;
		}

		networkRoles.splice(i, 1);
		instance.network_roles = networkRoles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addNetworkRole: '',
			instance: instance,
		});
	}

	onAddDriveDevice = (): void => {
		let instance: InstanceTypes.Instance;
		let infoDriveDevices = this.props.instance.info.drive_devices || [];

		if (!this.state.addDriveDevice && !infoDriveDevices.length) {
			return;
		}

		let addDevice = this.state.addDriveDevice;
		if (!addDevice) {
			addDevice = infoDriveDevices[0].id;
		}

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		let driveDevices = [
			...(instance.drive_devices || []),
		];

		let index = -1;
		for (let i = 0; i < driveDevices.length; i++) {
			let dev = driveDevices[i];
			if (dev.id === addDevice) {
				index = i;
				break
			}
		}

		if (index === -1) {
			driveDevices.push({
				id: addDevice,
			});
		}

		instance.drive_devices = driveDevices;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addDriveDevice: '',
			instance: instance,
		});
	}

	onRemoveDriveDevice = (device: string): void => {
		let instance: InstanceTypes.Instance;

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		let driveDevices = [
			...(instance.drive_devices || []),
		];

		let index = -1;
		for (let i = 0; i < driveDevices.length; i++) {
			let dev = driveDevices[i];
			if (dev.id === device) {
				index = i;
				break
			}
		}
		if (index === -1) {
			return;
		}

		driveDevices.splice(index, 1);
		instance.drive_devices = driveDevices;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addDriveDevice: '',
			instance: instance,
		});
	}

	onAddUsbDevice = (): void => {
		let instance: InstanceTypes.Instance;
		let infoUsbDevices = this.props.instance.info.usb_devices || [];

		if (!this.state.addUsbDevice && !infoUsbDevices.length) {
			return;
		}

		let addDevice = this.state.addUsbDevice;
		if (!addDevice) {
			addDevice = infoUsbDevices[0].vendor + ':' + infoUsbDevices[0].product;
		}

		let bus = addDevice.indexOf('-') !== -1;

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		let usbDevices = [
			...(instance.usb_devices || []),
		];

		let index = -1;
		for (let i = 0; i < usbDevices.length; i++) {
			let dev = usbDevices[i];
			if (!bus && dev.vendor + ':' + dev.product === addDevice) {
				index = i;
				break
			} else if (bus && dev.bus + '-' + dev.address === addDevice) {
				index = i;
				break
			}
		}

		if (!bus) {
			let device = addDevice.split(':');

			if (index === -1) {
				usbDevices.push({
					vendor: device[0],
					product: device[1],
				});
			}
		} else {
			let port = addDevice.split('-');

			if (index === -1) {
				usbDevices.push({
					bus: port[0],
					address: port[1],
				});
			}
		}

		instance.usb_devices = usbDevices;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addUsbDevice: '',
			instance: instance,
		});
	}

	onRemoveUsbDevice = (device: string): void => {
		let instance: InstanceTypes.Instance;

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		let usbDevices = [
			...(instance.usb_devices || []),
		];

		let bus = device.indexOf('-') !== -1;

		let index = -1;
		for (let i = 0; i < usbDevices.length; i++) {
			let dev = usbDevices[i];
			if (!bus && dev.vendor + ':' + dev.product == device) {
				index = i;
				break
			} else if (bus && dev.bus + '-' + dev.address == device) {
				index = i;
				break
			}
		}
		if (index === -1) {
			return;
		}

		usbDevices.splice(index, 1);
		instance.usb_devices = usbDevices;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addUsbDevice: '',
			instance: instance,
		});
	}

	onAddPciDevice = (): void => {
		let instance: InstanceTypes.Instance;
		let infoPciDevices = this.props.instance.info.pci_devices || [];

		if (!this.state.addPciDevice && !infoPciDevices.length) {
			return;
		}

		let addDevice = this.state.addPciDevice;
		if (!addDevice) {
			addDevice = infoPciDevices[0].slot;
		}

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		let pciDevices = [
			...(instance.pci_devices || []),
		];

		let index = -1;
		for (let i = 0; i < pciDevices.length; i++) {
			let dev = pciDevices[i];
			if (dev.slot === addDevice) {
				index = i;
				break
			}
		}

		if (index === -1) {
			pciDevices.push({
				slot: addDevice,
			});
		}

		instance.pci_devices = pciDevices;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addPciDevice: '',
			instance: instance,
		});
	}

	onRemovePciDevice = (device: string): void => {
		let instance: InstanceTypes.Instance;

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		let pciDevices = [
			...(instance.pci_devices || []),
		];

		let index = -1;
		for (let i = 0; i < pciDevices.length; i++) {
			let dev = pciDevices[i];
			if (dev.slot === device) {
				index = i;
				break
			}
		}
		if (index === -1) {
			return;
		}

		pciDevices.splice(index, 1);
		instance.pci_devices = pciDevices;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addPciDevice: '',
			instance: instance,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		InstanceActions.commit({
			...this.state.instance,
			state: null,
		}).then((): void => {
			this.setState({
				...this.state,
				message: 'Your changes have been saved',
				changed: false,
				disabled: false,
			});

			setTimeout((): void => {
				if (!this.state.changed) {
					this.setState({
						...this.state,
						instance: null,
						changed: false,
					});
				}
			}, 1000);

			setTimeout((): void => {
				if (!this.state.changed) {
					this.setState({
						...this.state,
						message: '',
					});
				}
			}, 3000);
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		InstanceActions.remove(this.props.instance.id).then((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	update(state: string): void {
		this.setState({
			...this.state,
			disabled: true,
		});
		InstanceActions.updateMulti([this.props.instance.id],
				state).then((): void => {
			setTimeout((): void => {
				this.setState({
					...this.state,
					disabled: false,
				});
			}, 250);
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let instance: InstanceTypes.Instance = this.state.instance ||
			this.props.instance;
		let info: InstanceTypes.Info = this.props.instance.info || {};

		let org = OrganizationsStore.organization(
			this.props.instance.organization);
		let zone = ZonesStore.zone(this.props.instance.zone);

		let privateIps: any = this.props.instance.private_ips;
		if (!privateIps || !privateIps.length) {
			privateIps = 'None';
		}

		let privateIps6: any = this.props.instance.private_ips6;
		if (!privateIps6 || !privateIps6.length) {
			privateIps6 = 'None';
		}

		let publicIps: any = this.props.instance.public_ips;
		if (!publicIps || !publicIps.length) {
			publicIps = 'None';
		}

		let publicIps6: any = this.props.instance.public_ips6;
		if (!publicIps6 || !publicIps6.length) {
			publicIps6 = 'None';
		}

		let hostIps: any = this.props.instance.host_ips;
		if (!hostIps || !hostIps.length) {
			hostIps = 'None';
		}

		let statusClass = '';
		switch (instance.status) {
			case 'Running':
				statusClass += 'bp3-text-intent-success';
				break;
			case 'Restart Required':
				statusClass += ' bp3-text-intent-warning';
				break;
			case 'Stopped':
			case 'Failed':
			case 'Destroying':
				statusClass += 'bp3-text-intent-danger';
				break;
		}

		let networkRoles: JSX.Element[] = [];
		for (let networkRole of (instance.network_roles || [])) {
			networkRoles.push(
				<div
					className="bp3-tag bp3-tag-removable bp3-intent-primary"
					style={css.role}
					key={networkRole}
				>
					{networkRole}
					<button
						className="bp3-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveNetworkRole(networkRole);
						}}
					/>
				</div>,
			);
		}

		let hasVpcs = false;
		let vpcsSelect: JSX.Element[] = [];
		if (this.props.vpcs && this.props.vpcs.length) {
			vpcsSelect.push(<option key="null" value="">Select Vpc</option>);

			for (let vpc of this.props.vpcs) {
				if (vpc.organization !== instance.organization) {
					continue;
				}

				hasVpcs = true;
				vpcsSelect.push(
					<option
						key={vpc.id}
						value={vpc.id}
					>{vpc.name}</option>,
				);
			}
		}

		if (!hasVpcs) {
			vpcsSelect = [<option key="null" value="">No Vpcs</option>];
		}

		let hasSubnets = false;
		let subnetSelect: JSX.Element[] = [];
		if (this.props.vpcs && this.props.vpcs.length) {
			subnetSelect.push(<option key="null" value="">Select Subnet</option>);

			for (let vpc of this.props.vpcs) {
				if (vpc.organization !== instance.organization) {
					continue;
				}

				if (vpc.id === instance.vpc) {
					for (let sub of (vpc.subnets || [])) {
						hasSubnets = true;
						subnetSelect.push(
							<option
								key={sub.id}
								value={sub.id}
							>{sub.name + ' - ' + sub.network}</option>,
						);
					}
				}
			}
		}

		if (!hasSubnets) {
			subnetSelect = [<option key="null" value="">No Subnets</option>];
		}

		let domainsSelect: JSX.Element[] = [
			<option key="null" value="">No Domain</option>,
		];
		if (this.props.domains && this.props.domains.length) {
			for (let domain of this.props.domains) {
				if (domain.organization !== instance.organization) {
					continue;
				}

				domainsSelect.push(
					<option
						key={domain.id}
						value={domain.id}
					>{domain.name}</option>,
				);
			}
		}

		let driveDevices: JSX.Element[] = [];
		for (let device of (instance.drive_devices || [])) {
			let key = device.id;
			driveDevices.push(
				<div
					className="bp3-tag bp3-tag-removable bp3-intent-primary"
					style={css.item}
					key={key}
				>
					{key}
					<button
						disabled={this.state.disabled}
						className="bp3-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveDriveDevice(key);
						}}
					/>
				</div>,
			);
		}

		let infoDriveDevices = this.props.instance.info.drive_devices;
		let driveDevicesSelect: JSX.Element[] = [];
		for (let i = 0; i < (infoDriveDevices || []).length; i++) {
			let device = infoDriveDevices[i];
			driveDevicesSelect.push(
				<option
					key={device.id}
					value={device.id}
				>
					{device.id}
				</option>,
			);
		}

		let usbDevices: JSX.Element[] = [];
		for (let device of (instance.usb_devices || [])) {
			let key = '';
			if (device.bus && device.address) {
				key = device.bus + '-' + device.address;
			} else {
				key = device.vendor + ':' + device.product;
			}
			usbDevices.push(
				<div
					className="bp3-tag bp3-tag-removable bp3-intent-primary"
					style={css.item}
					key={key}
				>
					{key}
					<button
						disabled={this.state.disabled}
						className="bp3-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveUsbDevice(key);
						}}
					/>
				</div>,
			);
		}

		let infoUsbDevices = this.props.instance.info.usb_devices;
		let usbDevicesSelect: JSX.Element[] = [];
		for (let i = 0; i < (infoUsbDevices || []).length; i++) {
			let device = infoUsbDevices[i];
			usbDevicesSelect.push(
				<option
					key={i + '_' + device.vendor + ':' + device.product}
					value={device.vendor + ':' + device.product}
				>
					{'Device=' + device.vendor + ':' + device.product +
					' (' + device.name + ')'}
				</option>,
			);
		}
		for (let i = 0; i < (infoUsbDevices || []).length; i++) {
			let device = infoUsbDevices[i];
			usbDevicesSelect.push(
				<option
					key={i + '_' + device.bus + '-' + device.address}
					value={device.bus + '-' + device.address}
				>
					{'Bus=' + device.bus + ' Port=' + device.address +
					' (' + device.name + ')'}
				</option>,
			);
		}

		let pciDevices: JSX.Element[] = [];
		for (let device of (instance.pci_devices || [])) {
			let key = device.slot;
			pciDevices.push(
				<div
					className="bp3-tag bp3-tag-removable bp3-intent-primary"
					style={css.item}
					key={key}
				>
					{key}
					<button
						disabled={this.state.disabled}
						className="bp3-tag-remove"
						onMouseUp={(): void => {
							this.onRemovePciDevice(key);
						}}
					/>
				</div>,
			);
		}

		let infoPciDevices = this.props.instance.info.pci_devices;
		let pciDevicesSelect: JSX.Element[] = [];
		for (let i = 0; i < (infoPciDevices || []).length; i++) {
			let device = infoPciDevices[i];
			pciDevicesSelect.push(
				<option
					key={device.slot}
					value={device.slot}
				>
					{device.slot + ' ' + device.class + ':' + device.name}
				</option>,
			);
		}

		let fields: PageInfos.Field[] = [
			{
				label: 'ID',
				value: this.props.instance.id || 'None',
			},
			{
				label: 'Organization',
				value: org ? org.name :
					this.props.instance.organization || 'None',
			},
			{
				label: 'Zone',
				value: zone ? zone.name : this.props.instance.zone || 'None',
			},
			{
				label: 'Node',
				value: info.node || 'None',
			},
			{
				label: 'State',
				value: (this.props.instance.state || 'None') + ':' + (
					this.props.instance.vm_state || 'None'),
			},
			{
				label: 'Uptime',
				value: this.props.instance.uptime || '-',
			},
			{
				label: 'Public MAC Address',
				value: this.props.instance.public_mac || 'Unknown',
				copy: true,
			},
			{
				label: 'Platform',
				value: (this.props.instance.uefi ? 'UEFI' : 'BIOS'),
			},
			{
				label: 'Public IPv4',
				value: publicIps,
				copy: true,
			},
			{
				label: 'Public IPv6',
				value: publicIps6,
				copy: true,
			},
			{
				label: 'Private IPv4',
				value: privateIps,
				copy: true,
			},
			{
				label: 'Private IPv6',
				value: privateIps6,
				copy: true,
			},
			{
				label: 'Host IPv4',
				value: hostIps,
				copy: true,
			},
			{
				label: 'Network Namespace',
				value: this.props.instance.network_namespace || 'None',
				copy: true,
			},
		];

		if (this.props.instance.vnc) {
			fields.push(
				{
					label: 'VNC Port',
					value: this.props.instance.vnc_display + 5900,
					copy: true,
				},
				{
					label: 'VNC Password',
					value: this.props.instance.vnc_password,
					copy: true,
				},
			);
		}

		fields.push(
			{
				label: 'Disks',
				value: info.disks || '',
			},
			{
				label: 'Firewall Rules',
				value: this.props.instance.info.firewall_rules || '',
			},
			{
				label: 'Authorities',
				value: this.props.instance.info.authorities || '',
			},
		);

		let vncStyle = {
			height: this.state.vncHeight ? this.state.vncHeight + 'px' : '100%',
		} as React.CSSProperties;

		return <td
			className="bp3-cell"
			colSpan={6}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal"
						style={css.buttons}
						onClick={(evt): void => {
							let target = evt.target as HTMLElement;

							if (target.className.indexOf('open-ignore') !== -1) {
								return;
							}

							this.props.onClose();
						}}
					>
            <div>
              <label
                className="bp3-control bp3-checkbox open-ignore"
                style={css.select}
              >
                <input
                  type="checkbox"
                  className="open-ignore"
                  checked={this.props.selected}
                  onClick={(evt): void => {
										this.props.onSelect(evt.shiftKey);
									}}
                />
                <span className="bp3-control-indicator open-ignore"/>
              </label>
            </div>
						<div className={statusClass} style={css.status}>
							<span
								style={css.icon}
								hidden={!instance.status}
								className="bp3-icon-standard bp3-icon-power"
							/>
							{instance.status}
						</div>
						<div className="flex"/>
						<ConfirmButton
							className="bp3-minimal bp3-intent-danger bp3-icon-trash open-ignore"
							style={css.button}
							progressClassName="bp3-intent-danger"
							confirmMsg="Confirm instance remove"
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of instance"
						type="text"
						placeholder="Enter name"
						value={instance.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageTextArea
						label="Comment"
						help="Instance comment."
						placeholder="Instance comment"
						rows={3}
						value={instance.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
					<PageNumInput
						label="Memory Size"
						help="Instance memory size in megabytes."
						min={256}
						minorStepSize={256}
						stepSize={512}
						majorStepSize={1024}
						disabled={this.state.disabled}
						selectAllOnFocus={true}
						onChange={(val: number): void => {
							this.set('memory', val);
						}}
						value={instance.memory}
					/>
					<PageNumInput
						label="Processors"
						help="Number of instance processors."
						min={1}
						minorStepSize={1}
						stepSize={1}
						majorStepSize={2}
						disabled={this.state.disabled}
						selectAllOnFocus={true}
						onChange={(val: number): void => {
							this.set('processors', val);
						}}
						value={instance.processors}
					/>
					<label className="bp3-label">
						Network Roles
						<Help
							title="Network Roles"
							content="Network roles that will be matched with firewall rules. Network roles are case-sensitive."
						/>
						<div>
							{networkRoles}
						</div>
					</label>
					<PageInputButton
						disabled={this.state.disabled}
						buttonClass="bp3-intent-success bp3-icon-add"
						label="Add"
						type="text"
						placeholder="Add role"
						value={this.state.addNetworkRole}
						onChange={(val): void => {
							this.setState({
								...this.state,
								addNetworkRole: val,
							});
						}}
						onSubmit={this.onAddNetworkRole}
					/>
					<label
						className="bp3-label"
						style={css.label}
						hidden={infoDriveDevices === null}
					>
						Disk Passthrough Devices
						<Help
							title="Disk Passthrough Devices"
							content="Passthrough node disk to instance."
						/>
						<div>
							{driveDevices}
						</div>
					</label>
					<PageSelectButton
						hidden={infoDriveDevices === null}
						label="Add Device"
						value={this.state.addDriveDevice}
						disabled={!driveDevicesSelect.length || this.state.disabled}
						buttonClass="bp3-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								addDriveDevice: val,
							});
						}}
						onSubmit={this.onAddDriveDevice}
					>
						{driveDevicesSelect}
					</PageSelectButton>
					<label
						className="bp3-label"
						style={css.label}
						hidden={infoPciDevices === null}
					>
						PCI Devices
						<Help
							title="PCI Devices"
							content="PCI devices to for host passthrough to instance."
						/>
						<div>
							{pciDevices}
						</div>
					</label>
					<PageSelectButton
						hidden={infoPciDevices === null}
						label="Add Device"
						value={this.state.addPciDevice}
						disabled={!pciDevicesSelect.length || this.state.disabled}
						buttonClass="bp3-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								addPciDevice: val,
							});
						}}
						onSubmit={this.onAddPciDevice}
					>
						{pciDevicesSelect}
					</PageSelectButton>
					<label
						className="bp3-label"
						style={css.label}
						hidden={infoUsbDevices === null}
					>
						USB Devices
						<Help
							title="USB Devices"
							content="USB devices to for host passthrough to instance."
						/>
						<div>
							{usbDevices}
						</div>
					</label>
					<PageSelectButton
						hidden={infoUsbDevices === null}
						label="Add Device"
						value={this.state.addUsbDevice}
						disabled={!usbDevicesSelect.length || this.state.disabled}
						buttonClass="bp3-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								addUsbDevice: val,
							});
						}}
						onSubmit={this.onAddUsbDevice}
					>
						{usbDevicesSelect}
					</PageSelectButton>
					<PageSelect
						disabled={this.state.disabled || !hasVpcs}
						label="VPC"
						help="VPC for instance."
						value={instance.vpc}
						onChange={(val): void => {
							this.set('vpc', val);
						}}
					>
						{vpcsSelect}
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled || !hasVpcs}
						label="Subnet"
						help="Subnet for instance."
						value={instance.subnet}
						onChange={(val): void => {
							this.set('subnet', val);
						}}
					>
						{subnetSelect}
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled}
						label="DNS Domain"
						help="Domain to create DNS name using instance name."
						value={instance.domain}
						onChange={(val): void => {
							this.set('domain', val);
						}}
					>
						{domainsSelect}
					</PageSelect>
					<PageSwitch
						disabled={this.state.disabled}
						label="VNC server"
						help="Enable VNC server for remote control of instance."
						checked={instance.vnc}
						onToggle={(): void => {
							this.set('vnc', !instance.vnc);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="UEFI"
						help="Enable UEFI boot, requires UEFI image."
						checked={instance.uefi}
						onToggle={(): void => {
							this.set('uefi', !instance.uefi);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Delete protection"
						help="Block instance and any attached disks from being deleted."
						checked={instance.delete_protection}
						onToggle={(): void => {
							this.set('delete_protection', !instance.delete_protection);
						}}
					/>
					<PageSwitch
						label="Public address"
						help="Enable or disable public address for instance. Node must have network mode configured to assign public address."
						checked={!instance.no_public_address}
						onToggle={(): void => {
							this.set('no_public_address', !instance.no_public_address);
						}}
					/>
					<PageSwitch
						label="Host address"
						help="Enable or disable host address for instance. Node must have host networking configured to assign host address."
						checked={!instance.no_host_address}
						onToggle={(): void => {
							this.set('no_host_address', !instance.no_host_address);
						}}
					/>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={fields}
					/>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.instance && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						forwardedChecked: false,
						instance: null,
					});
				}}
				onSave={this.onSave}
			>
				<ConfirmButton
					label="Start"
					className="bp3-intent-success bp3-icon-power"
					progressClassName="bp3-intent-success"
					style={css.controlButton}
					hidden={this.props.instance.state !== 'stop'}
					disabled={this.state.disabled}
					onConfirm={(): void => {
						this.update('start');
					}}
				/>
				<ConfirmButton
					label="Stop"
					className="bp3-intent-danger bp3-icon-power"
					progressClassName="bp3-intent-danger"
					style={css.controlButton}
					hidden={this.props.instance.state !== 'start'}
					disabled={this.state.disabled}
					onConfirm={(): void => {
						this.update('stop');
					}}
				/>
				<button
					className="bp3-button bp3-intent-success bp3-icon-console"
					hidden={this.state.vnc || !this.props.instance.vnc}
					style={css.controlButton}
					disabled={this.state.disabled}
					type="button"
					onClick={(): void => {
						this.onTogleVnc();
					}}
				>
					VNC Console
				</button>
				<button
					className="bp3-button bp3-intent-danger bp3-icon-console"
					hidden={!this.state.vnc}
					style={css.controlButton}
					disabled={this.state.disabled}
					type="button"
					onClick={(): void => {
						this.onTogleVnc();
					}}
				>
					VNC Console
				</button>
				<button
					className={'bp3-button bp3-icon-key-control' +
						(this.state.vncCtrl ? ' bp3-active' : '')}
					hidden={!this.state.vnc}
					style={css.controlButton}
					disabled={this.state.disabled}
					type="button"
					onClick={(): void => {
						this.onToggleVncCtrl();
					}}
				>
					Ctrl
				</button>
				<button
					className={'bp3-button bp3-icon-key-option' +
						(this.state.vncAlt ? ' bp3-active' : '')}
					hidden={!this.state.vnc}
					style={css.controlButton}
					disabled={this.state.disabled}
					type="button"
					onClick={(): void => {
						this.onToggleVncAlt();
					}}
				>
					Alt
				</button>
				<button
					className={'bp3-button bp3-icon-key-command' +
						(this.state.vncSuper ? ' bp3-active' : '')}
					hidden={!this.state.vnc}
					style={css.controlButton}
					disabled={this.state.disabled}
					type="button"
					onClick={(): void => {
						this.onToggleVncSuper();
					}}
				>
					Super
				</button>
				<button
					className="bp3-button bp3-icon-key-tab"
					hidden={!this.state.vnc}
					style={css.controlButton}
					disabled={this.state.disabled}
					type="button"
					onClick={(): void => {
						this.onVncTab();
					}}
				>
					Tab
				</button>
				<button
					className="bp3-button bp3-icon-key-escape"
					hidden={!this.state.vnc}
					style={css.controlButton}
					disabled={this.state.disabled}
					type="button"
					onClick={(): void => {
						this.onVncEsc();
					}}
				>
					Esc
				</button>
				<button
					className="bp3-button bp3-icon-fullscreen"
					hidden={!this.state.vnc}
					style={css.controlButton}
					disabled={this.state.disabled}
					type="button"
					onClick={(): void => {
						this.onToggleVncFullscreen();
					}}
				>
					Fullscreen
				</button>
				<button
					className={'bp3-button bp3-icon-zoom-to-fit' +
						(this.state.vncScale ? ' bp3-active' : '')}
					hidden={!this.state.vnc}
					style={css.controlButton}
					disabled={this.state.disabled}
					type="button"
					onClick={(): void => {
						this.onToggleVncScale();
					}}
				>
					Scale
				</button>
				<button
					className="bp3-button bp3-intent-danger bp3-icon-control"
					hidden={!this.state.vnc}
					disabled={this.state.disabled}
					type="button"
					onClick={(): void => {
						this.onVncCtrlAltDel();
					}}
				>
					Ctrl-Alt-Del
				</button>
			</PageSave>
			<div style={css.vncBox}>
				<div
					ref={this.vncRef}
					style={vncStyle}
					hidden={!this.state.vnc}
				/>
			</div>
		</td>;
	}
}
