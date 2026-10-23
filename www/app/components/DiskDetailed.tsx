/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from "@blueprintjs/core";
import * as Theme from '../Theme';
import * as DiskTypes from '../types/DiskTypes';
import * as DiskActions from '../actions/DiskActions';
import * as OrganizationTypes from '../types/OrganizationTypes';
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageSwitch from './PageSwitch';
import PageNumInput from './PageNumInput';
import PageInfo from './PageInfo';
import PageSelectButtonConfirm from './PageSelectButtonConfirm';
import Help from './Help';
import * as PageInfos from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import CompletionStore from '../stores/CompletionStore';
import Relations from './Relations';
import * as InstanceActions from '../actions/InstanceActions';
import InstancesNodeStore from '../stores/InstancesNodeStore';
import * as InstanceTypes from '../types/InstanceTypes';
import * as Alert from '../Alert';
import PageTextArea from "./PageTextArea";
import * as PoolTypes from "../types/PoolTypes";
import * as MiscUtils from '../utils/MiscUtils';

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	pools: PoolTypes.PoolsRo;
	disk: DiskTypes.DiskRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	disk: DiskTypes.Disk;
	instances: InstanceTypes.InstancesRo;
	restoreImage: string;
	resizeDisk: boolean;
}

interface FsSegment {
	key: string;
	label: string;
	detail: string;
	title: string;
	size: number;
	fileSystem?: string;
	free?: boolean;
}

const fsMinGap = 16 * 1024 * 1024;
const fsColorDefault = '143, 153, 168';
const fsColors: {[key: string]: string} = {
	'xfs': '45, 114, 210',
	'ext4': '35, 133, 81',
	'ext3': '35, 133, 81',
	'ext2': '35, 133, 81',
	'btrfs': '0, 153, 150',
	'vfat': '125, 88, 192',
	'swap': '171, 83, 160',
	'LVM2_member': '200, 118, 25',
};

function fsName(fileSystem: string): string {
	if (!fileSystem) {
		return 'unformatted';
	}
	if (fileSystem === 'LVM2_member') {
		return 'lvm';
	}
	return fileSystem;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 0 10px',
		width: '100%',
	} as React.CSSProperties,
	button: {
		height: '30px',
	} as React.CSSProperties,
	buttons: {
		cursor: 'pointer',
		position: 'absolute',
		top: 0,
		left: 0,
		right: 0,
		padding: '4px',
		height: '39px',
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		minHeight: '20px',
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
	controlButton: {
		marginRight: '10px',
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
		minHeight: '20px',
	} as React.CSSProperties,
	rules: {
		marginBottom: '15px',
	} as React.CSSProperties,
	instSection: {
		width: '100%',
		padding: '0 10px',
		marginTop: '4px',
	} as React.CSSProperties,
	section: {
		display: 'flex',
		alignItems: 'center',
		margin: '14px 0 8px 0',
		fontWeight: 600,
	} as React.CSSProperties,
	sectionIcon: {
		marginRight: '6px',
	} as React.CSSProperties,
	indexTag: {
		marginLeft: '8px',
		paddingTop: '3px',
		paddingBottom: '3px',
	} as React.CSSProperties,
	instCard: {
		padding: '5px 0 0 0',
		marginBottom: '10px',
	} as React.CSSProperties,
	instBox: {
		fontSize: '11px',
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
	instInfo: {
		marginBottom: '0px',
		fontSize: '11px',
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
	instItemFirst: {
		flex: '1 1 auto',
		minWidth: '100px',
		maxWidth: '170px',
		margin: '0 5px 0 10px',
	} as React.CSSProperties,
	instItem: {
		flex: '1 1 auto',
		width: 0,
		margin: '0 5px',
	} as React.CSSProperties,
	instItemFull: {
		flex: '2 1 auto',
		width: 0,
		margin: '0 5px',
	} as React.CSSProperties,
	fsScanned: {
		fontSize: '11px',
		fontWeight: 400,
	} as React.CSSProperties,
	fsCard: {
		padding: '8px 10px 6px 10px',
		marginBottom: '10px',
	} as React.CSSProperties,
	fsBlock: {
		marginBottom: '10px',
	} as React.CSSProperties,
	fsBarHeader: {
		alignItems: 'baseline',
		marginBottom: '3px',
		fontSize: '11px',
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
	fsBar: {
		display: 'flex',
		width: '100%',
		gap: '3px',
	} as React.CSSProperties,
	fsSegment: {
		minWidth: '60px',
		height: '38px',
		padding: '3px 6px',
		overflow: 'hidden',
		boxSizing: 'border-box',
		borderRadius: '3px',
		fontSize: '11px',
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
		lineHeight: '15px',
	} as React.CSSProperties,
	fsSegmentFree: {
		border: '1px dashed rgba(143, 153, 168, 0.6)',
	} as React.CSSProperties,
	fsSegmentLine: {
		overflow: 'hidden',
		whiteSpace: 'nowrap',
		textOverflow: 'ellipsis',
	} as React.CSSProperties,
	fsTable: {
		width: '100%',
		marginTop: '2px',
		fontSize: '11px',
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
	fsUuid: {
		wordBreak: 'break-all',
	} as React.CSSProperties,
};

export default class DiskDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			disk: null,
			instances: null,
			restoreImage: null,
			resizeDisk: false,
		};
	}

	componentDidMount(): void {
		InstancesNodeStore.addChangeListener(this.onChange);
		InstanceActions.syncNode(this.props.disk.node, this.props.disk.pool);
	}

	componentWillUnmount(): void {
		InstancesNodeStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			instances: InstancesNodeStore.instances(
				this.props.disk.node || this.props.disk.pool),
		});
	}

	set(name: string, val: any): void {
		let disk: any;

		if (this.state.changed) {
			disk = {
				...this.state.disk,
			};
		} else {
			disk = {
				...this.props.disk,
			};
		}

		disk[name] = val;

		if (name === 'instance' && !Number(disk.index)) {
			disk['index'] = '0';
		}

		this.setState({
			...this.state,
			changed: true,
			disk: disk,
		});
	}

	setResizeDisk(val: boolean): void {
		let disk: any;

		if (this.state.changed) {
			disk = {
				...this.state.disk,
			};
		} else {
			disk = {
				...this.props.disk,
			};
		}

		if (val) {
			disk.new_size = disk.size;
		} else {
			disk.new_size = 0;
		}

		this.setState({
			...this.state,
			changed: true,
			resizeDisk: val,
			disk: disk,
		});

	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let disk = {
			...this.state.disk,
		};

		if (this.state.resizeDisk && disk.new_size > disk.size) {
			disk.action = 'expand';
		}

		DiskActions.commit(disk).then((): void => {
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
						disk: null,
						changed: false,
						resizeDisk: false,
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
		DiskActions.remove(this.props.disk.id).then((): void => {
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

	onRestoreBackup = (): void => {
		let restoreImage: string;

		if (this.state.restoreImage) {
			restoreImage = this.state.restoreImage;
		} else if (this.props.disk.backups && this.props.disk.backups.length) {
			restoreImage = this.props.disk.backups[0].image;
		} else {
			return;
		}

		this.setState({
			...this.state,
			disabled: true,
		});

		let disk: DiskTypes.Disk;

		if (this.state.changed) {
			disk = {
				...this.state.disk,
			};
		} else {
			disk = {
				...this.props.disk,
			};
		}

		disk.action = 'restore';
		disk.restore_image = restoreImage;

		DiskActions.commit(disk).then((): void => {
			Alert.success('Disk restore started');

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

	onAction = (action: string, message: string): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let disk: DiskTypes.Disk = {
			...this.props.disk,
		};
		disk.action = action;

		DiskActions.commit(disk).then((): void => {
			Alert.success(message);

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

	instanceStatusClass(status: string): string {
		switch (status) {
			case 'Running':
				return 'bp5-text-intent-success';
			case 'Starting':
			case 'Restarting':
			case 'Updating':
			case 'Provisioning':
				return 'bp5-text-intent-primary';
			case 'Failed':
			case 'Stopping':
			case 'Stopped':
			case 'Destroying':
				return 'bp5-text-intent-danger';
			default:
				if (status && status.startsWith('Restart Required')) {
					return 'bp5-text-intent-warning';
				}
				return '';
		}
	}

	renderInstanceCard(): JSX.Element {
		let inst = this.props.disk.instance_info;
		if (!inst) {
			return null;
		}

		let statusValue = inst.status || '-';
		let statusClass = this.instanceStatusClass(inst.status || '');

		let publicIps = inst.public_ips && inst.public_ips.length ?
			inst.public_ips : ['-'];
		let publicIps6 = inst.public_ips6 && inst.public_ips6.length ?
			inst.public_ips6 : ['-'];
		let privateIps = inst.private_ips && inst.private_ips.length ?
			inst.private_ips : ['-'];
		let privateIps6 = inst.private_ips6 && inst.private_ips6.length ?
			inst.private_ips6 : ['-'];

		let index = this.props.disk.index;

		return <div style={css.instSection}>
			<div style={css.section}>
				<span
					className="bp5-icon-standard bp5-icon-data-connection"
					style={css.sectionIcon}
				/>
				Attached Instance
				{Number(index) || index === '0' ? <Blueprint.Tag
					minimal={true}
					style={css.indexTag}
				>Index {index}</Blueprint.Tag> : null}
			</div>
			<Blueprint.Card
				compact={true}
				style={css.instCard}
			>
				<div className="layout horizontal flex" style={css.instBox}>
					<div style={css.instItemFirst}>
						<PageInfo
							compact={true}
							style={css.instInfo}
							fields={[
								{
									label: 'Instance',
									value: inst.name || '-',
								},
								{
									label: 'Instance ID',
									value: inst.id || '-',
								},
							]}
						/>
					</div>
					<div style={css.instItem}>
						<PageInfo
							compact={true}
							style={css.instInfo}
							fields={[
								{
									label: 'Status',
									value: statusValue,
									valueClass: statusClass,
								},
								{
									label: 'Uptime',
									value: inst.uptime || '-',
								},
							]}
						/>
					</div>
					<div style={css.instItem}>
						<PageInfo
							compact={true}
							style={css.instInfo}
							fields={[
								{
									label: 'Private IPv4',
									value: privateIps,
								},
								{
									label: 'Public IPv4',
									value: publicIps,
								},
							]}
						/>
					</div>
					<div style={css.instItemFull}>
						<PageInfo
							compact={true}
							style={css.instInfo}
							fields={[
								{
									label: 'Private IPv6',
									value: privateIps6,
									maxLines: 2,
								},
								{
									label: 'Public IPv6',
									value: publicIps6,
									maxLines: 2,
								},
							]}
						/>
					</div>
				</div>
			</Blueprint.Card>
		</div>;
	}

	renderFsSegments(segments: FsSegment[], total: number): JSX.Element {
		let elems: JSX.Element[] = [];

		for (let seg of segments) {
			let percent = total > 0 ? (seg.size / total) * 100 : 0;
			let style: React.CSSProperties = {
				...css.fsSegment,
				flex: percent.toFixed(4) + ' 1 0px',
			};

			if (seg.free) {
				style = {
					...style,
					...css.fsSegmentFree,
				};
			} else {
				let color = fsColors[seg.fileSystem] || fsColorDefault;
				style.backgroundColor = 'rgba(' + color + ', 0.22)';
				style.border = '1px solid rgba(' + color + ', 0.75)';
			}

			elems.push(
				<div key={seg.key} style={style} title={seg.title}>
					<div style={css.fsSegmentLine}>{seg.label}</div>
					<div
						className="bp5-text-muted"
						style={css.fsSegmentLine}
					>{seg.detail}</div>
				</div>,
			);
		}

		return <div style={css.fsBar}>{elems}</div>;
	}

	renderFsBar(key: string, label: string, detail: string,
		segments: FsSegment[], total: number): JSX.Element {

		return <div key={key} style={css.fsBlock}>
			<div className="layout horizontal" style={css.fsBarHeader}>
				<div>{label}</div>
				<div className="flex"/>
				<div className="bp5-text-muted">{detail}</div>
			</div>
			{this.renderFsSegments(segments, total)}
		</div>;
	}

	renderFsRow(key: string, name: string, fileSystem: string,
		label: string, size: number, uuid: string): JSX.Element {

		return <tr key={key}>
			<td>{name}</td>
			<td>{fileSystem}</td>
			<td>{label || '-'}</td>
			<td>{MiscUtils.humanReadableBytes(size)}</td>
			<td style={css.fsUuid}>{uuid || '-'}</td>
		</tr>;
	}

	renderFileSystem(): JSX.Element {
		let info = this.props.disk.metadata;
		if (!info || !info.size) {
			return null;
		}

		let partitions = (info.partitions || []).slice().sort(
			(a, b): number => (a.start || 0) - (b.start || 0));

		let diskSegments: FsSegment[] = [];
		let vgBars: JSX.Element[] = [];
		let rows: JSX.Element[] = [];
		let pos = 0;

		let addFree = (start: number, size: number): void => {
			if (size < fsMinGap) {
				return;
			}

			let sizeStr = MiscUtils.humanReadableBytes(size);
			diskSegments.push({
				key: 'free-' + start,
				label: 'Unallocated',
				detail: sizeStr,
				title: 'Unallocated ' + sizeStr,
				size: size,
				free: true,
			});
			rows.push(this.renderFsRow('free-' + start, 'Unallocated',
				'-', '', size, ''));
		};

		for (let part of partitions) {
			let start = part.start || 0;
			let size = part.size || 0;
			let key = 'part-' + part.number + '-' + start;

			if (start > pos) {
				addFree(pos, start - pos);
			}

			let name = part.number ? 'Partition ' + part.number : 'Disk';
			let fs = fsName(part.file_system);
			let sizeStr = MiscUtils.humanReadableBytes(size);

			diskSegments.push({
				key: key,
				label: name,
				detail: fs + ' \u00b7 ' + sizeStr,
				title: name + '\n' + fs + ' ' + sizeStr +
					(part.label ? '\nLabel: ' + part.label : '') +
					(part.uuid ? '\nUUID: ' + part.uuid : ''),
				size: size,
				fileSystem: part.file_system,
			});
			rows.push(this.renderFsRow(key, name, fs, part.label,
				size, part.uuid));

			let vg = part.volume_group;
			if (vg) {
				let vgSegments: FsSegment[] = [];

				for (let lv of (vg.volumes || [])) {
					let lvName = vg.name + '/' + lv.name;
					let lvFs = fsName(lv.file_system);
					let lvSize = lv.size || 0;
					let lvSizeStr = MiscUtils.humanReadableBytes(lvSize);

					vgSegments.push({
						key: 'lv-' + lvName,
						label: lv.name,
						detail: lvFs + ' \u00b7 ' + lvSizeStr,
						title: 'Logical Volume ' + lvName + '\n' +
							lvFs + ' ' + lvSizeStr +
							(lv.label ? '\nLabel: ' + lv.label : '') +
							(lv.uuid ? '\nUUID: ' + lv.uuid : ''),
						size: lvSize,
						fileSystem: lv.file_system,
					});
					rows.push(this.renderFsRow('lv-' + lvName, lvName,
						lvFs, lv.label, lvSize, lv.uuid));
				}

				let vgFree = vg.free || 0;
				if (vgFree > 0) {
					let freeStr = MiscUtils.humanReadableBytes(vgFree);
					vgSegments.push({
						key: 'vg-free-' + vg.name,
						label: 'Free',
						detail: freeStr,
						title: 'Volume Group ' + vg.name +
							'\nFree ' + freeStr,
						size: vgFree,
						free: true,
					});
					rows.push(this.renderFsRow('vg-free-' + vg.name,
						vg.name + ' (free)', '-', '', vgFree, ''));
				}

				vgBars.push(this.renderFsBar('vg-' + vg.name,
					'Volume Group ' + vg.name,
					MiscUtils.humanReadableBytes(vg.size || 0) +
						' \u00b7 ' + MiscUtils.humanReadableBytes(vgFree) +
						' free',
					vgSegments, vg.size || 0));
			}

			pos = Math.max(pos, start + size);
		}

		if (info.size > pos) {
			addFree(pos, info.size - pos);
		}

		return <div style={css.instSection}>
			<div style={css.section}>
				<span
					className="bp5-icon-standard bp5-icon-database"
					style={css.sectionIcon}
				/>
				File System
				{info.table ? <Blueprint.Tag
					minimal={true}
					style={css.indexTag}
				>{info.table.toUpperCase()}</Blueprint.Tag> : null}
				<div className="flex"/>
				<span className="bp5-text-muted" style={css.fsScanned}>
					Scanned {MiscUtils.formatDateShortTime(info.timestamp)}
				</span>
			</div>
			<Blueprint.Card
				compact={true}
				style={css.fsCard}
			>
				{this.renderFsBar('disk', 'Disk',
					MiscUtils.humanReadableBytes(info.size),
					diskSegments, info.size)}
				{vgBars}
				<table
					className="bp5-html-table bp5-compact"
					style={css.fsTable}
				>
					<thead>
						<tr>
							<th>Partition</th>
							<th>File System</th>
							<th>Label</th>
							<th>Size</th>
							<th>UUID</th>
						</tr>
					</thead>
					<tbody>{rows}</tbody>
				</table>
			</Blueprint.Card>
		</div>;
	}

	render(): JSX.Element {
		let disk: DiskTypes.Disk = this.state.disk ||
			this.props.disk;

		let node = CompletionStore.node(this.props.disk.node);
		let org = CompletionStore.organization(this.props.disk.organization);

		let hasInstances = false;
		let instancesSelect: JSX.Element[] = [];
		if (this.state.instances === null) {
			instancesSelect = [<option key="null" value="">Loading</option>];
		} else {
			if (this.state.instances.length) {
				instancesSelect.push(
					<option key="null" value="">Detached Disk</option>);

				hasInstances = true;
				for (let instance of this.state.instances) {
					instancesSelect.push(
						<option
							key={instance.id}
							value={instance.id}
						>{instance.name}</option>,
					);
				}
			}

			if (!hasInstances) {
				instancesSelect = [<option key="null" value="">No Instances</option>];
			}
		}

		let resourceLabel = "";
		let resourceValue = "";
		if (disk.type === "lvm") {
			resourceLabel = "Pool"
			resourceValue = "Pool Unavailable"

			if (this.props.pools.length) {
				for (let pool of this.props.pools) {
					if (pool.id === disk.pool) {
						resourceValue = pool.name;
						break;
					}
				}
			}
		} else {
			resourceLabel = "Node";
			resourceValue = (node ? node.name : this.props.disk.node) || '-';
		}

		let backupsSelect: JSX.Element[] = [];
		for (let backup of (disk.backups || [])) {
			backupsSelect.push(
				<option key={backup.image} value={backup.image}>
					{backup.name}
				</option>,
			);
		}

		let hasBackups = false;
		if (!backupsSelect.length) {
			backupsSelect = [<option key="null" value="">No Backups</option>];
		} else {
			hasBackups = true
		}

		let statusText = 'Unknown';
		let statusClass = 'tab-close ';
		switch (this.props.disk.state) {
			case 'provision':
				statusText = 'Provisioning';
				statusClass += ' bp5-text-intent-primary';
				break;
			case 'available':
				if (this.props.disk.instance) {
					statusText = 'Connected';
				} else {
					statusText = 'Available';
				}
				statusClass += ' bp5-text-intent-success';
				break;
			case 'attached':
				statusText = 'Connected';
				statusClass += ' bp5-text-intent-success';
				break;
		}

		let actionLabel = 'Action';
		switch (this.props.disk.action) {
			case 'destroy':
				statusText = 'Destroying';
				statusClass += ' bp5-text-intent-danger';
				actionLabel = 'Destroy';
				break;
			case 'snapshot':
				statusText = 'Snapshotting';
				statusClass += ' bp5-text-intent-primary';
				actionLabel = 'Snapshot';
				break;
			case 'backup':
				statusText = 'Backing Up';
				statusClass += ' bp5-text-intent-primary';
				actionLabel = 'Backup';
				break;
			case 'restore':
				statusText = 'Restoring';
				statusClass += ' bp5-text-intent-primary';
				actionLabel = 'Restore';
				break;
			case 'expand':
				statusText = 'Expanding';
				statusClass += ' bp5-text-intent-primary';
				actionLabel = 'Expand';
				break;
		}

		let fields: PageInfos.Field[] = [
			{
				label: 'ID',
				value: this.props.disk.id || 'Unknown',
			},
			{
				label: 'Organization',
				value: org ? org.name : this.props.disk.organization,
			},
			{
				label: resourceLabel,
				value: resourceValue,
			},
			{
				label: 'Size',
				value: this.props.disk.size + 'GB',
			},
		];

		if (this.props.disk.uuid) {
			fields.splice(2, 0, {
				label: 'UUID',
				value: this.props.disk.uuid,
			});
		}

		if (this.props.disk.file_system) {
			fields.splice(3, 0, {
				label: 'File System',
				value: this.props.disk.file_system,
			});
		}

		if (this.props.disk.image) {
			fields.splice(2, 0, {
				label: 'Image',
				value: this.props.disk.image || 'Blank Disk',
			});
		}

		if (this.props.disk.deployment) {
			fields.push({
				label: 'Deployment',
				value: this.props.disk.deployment,
			})
		}

		let backingImage = this.props.disk.backing_image;
		if (backingImage) {
			backingImage = backingImage.replace('-', '\n');

			fields.splice(2, 0, {
				label: 'Backing Image',
				value: backingImage,
			});
		}

		return <td
			className="bp5-cell"
			colSpan={5}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal tab-close bp5-card-header"
						style={css.buttons}
						onClick={(evt): void => {
							if (evt.target instanceof HTMLElement &&
									evt.target.className.indexOf('tab-close') !== -1) {
								this.props.onClose();
							}
						}}
					>
            <div>
              <label
                className="bp5-control bp5-checkbox tab-close"
                style={css.select}
              >
                <input
                  type="checkbox"
                  checked={this.props.selected}
									onChange={(evt): void => {
									}}
                  onClick={(evt): void => {
										this.props.onSelect(evt.shiftKey);
									}}
                />
                <span className="bp5-control-indicator"/>
              </label>
            </div>
						<div className={statusClass} style={css.status}>
							<span
								style={css.icon}
								className="bp5-icon-standard bp5-icon-pulse"
							/>
							{statusText}
						</div>
						<div className="flex tab-close"/>
						<Relations kind="disk" id={this.props.disk.id}/>
						<ConfirmButton
							className="bp5-minimal bp5-intent-danger bp5-icon-trash"
							style={css.button}
							safe={true}
							progressClassName="bp5-intent-danger"
							dialogClassName="bp5-intent-danger bp5-icon-delete"
							dialogLabel="Delete Disk"
							confirmMsg="Permanently delete this disk"
							confirmInput={true}
							items={[disk.name]}
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of disk."
						type="text"
						placeholder="Enter name"
						value={disk.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageTextArea
						label="Comment"
						help="Disk comment."
						placeholder="Disk comment"
						rows={6}
						value={disk.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
					<PageSelect
						disabled={true}
						label="Type"
						help="Type of disk. QCOW disk files are stored locally on the node filesystem. LVM disks are partitioned as a logical volume."
						value={disk.type}
						onChange={(val): void => {
							this.set('type', val);
						}}
					>
						<option key="qcow2" value="qcow2">QCOW</option>
						<option key="lvm" value="lvm">LVM</option>
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled || !hasInstances}
						label="Instance"
						help="Instance to attach disk to."
						value={disk.instance}
						onChange={(val): void => {
							this.set('instance', val);
						}}
					>
						{instancesSelect}
					</PageSelect>
					<PageNumInput
						label="Index"
						help="Index to attach disk."
						hidden={!disk.instance}
						min={0}
						max={8}
						minorStepSize={1}
						stepSize={1}
						majorStepSize={1}
						disabled={this.state.disabled}
						selectAllOnFocus={true}
						value={Number(disk.index) || 0}
						onChange={(val: number): void => {
							this.set('index', String(val));
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Delete protection"
						help="Block disk from being deleted."
						checked={disk.delete_protection}
						onToggle={(): void => {
							this.set('delete_protection', !disk.delete_protection);
						}}
					/>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={fields}
					/>
					<PageSwitch
						disabled={this.state.disabled || disk.state != 'available'}
						label="Resize disk"
						help="Change size of disk. Instance will be stopped."
						checked={this.state.resizeDisk}
						onToggle={(): void => {
							this.setResizeDisk(!this.state.resizeDisk);
						}}
					/>
					<PageNumInput
						label="New Size"
						help="New disk size in gigabytes."
						hidden={!this.state.resizeDisk}
						min={disk.size}
						minorStepSize={1}
						stepSize={1}
						majorStepSize={1}
						disabled={this.state.disabled}
						selectAllOnFocus={true}
						value={disk.new_size}
						onChange={(val: number): void => {
							this.set('new_size', val);
						}}
					/>
					<PageSwitch
						disabled={this.state.disabled}
						label="Automatic backup"
						help="Automatically backup disk daily."
						checked={disk.backup}
						onToggle={(): void => {
							this.set('backup', !disk.backup);
						}}
					/>
					<label
						className="bp5-label"
						style={css.label}
					>
						Restore Backup
						<Help
							title="Restore Backup"
							content="Select a backup to restore and replace the existing disk with the backup image. Instance will be stopped."
						/>
					</label>
					<PageSelectButtonConfirm
						label="Restore"
						value={this.state.restoreImage}
						disabled={!hasBackups || this.state.disabled}
						confirmMsg="Confirm disk restore"
						buttonClass="bp5-intent-success bp5-icon-box"
						progressClassName="bp5-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								restoreImage: val,
							});
						}}
						onSubmit={this.onRestoreBackup}
					>
						{backupsSelect}
					</PageSelectButtonConfirm>
				</div>
			</div>
			{this.renderFileSystem()}
			{this.renderInstanceCard()}
			<PageSave
				style={css.save}
				hidden={!this.state.disk && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						disk: null,
					});
				}}
				onSave={this.onSave}
			>
				<ConfirmButton
					label="Snapshot"
					className="bp5-intent-primary bp5-icon-floppy-disk"
					progressClassName="bp5-intent-primary"
					style={css.controlButton}
					disabled={this.state.disabled || !!this.props.disk.action}
					onConfirm={(): void => {
						this.onAction('snapshot', 'Disk snapshot started');
					}}
				/>
				<ConfirmButton
					label="Backup"
					className="bp5-intent-primary bp5-icon-compressed"
					progressClassName="bp5-intent-primary"
					style={css.controlButton}
					disabled={this.state.disabled || !!this.props.disk.action}
					onConfirm={(): void => {
						this.onAction('backup', 'Disk backup started');
					}}
				/>
				<ConfirmButton
					label={'Cancel ' + actionLabel}
					className="bp5-intent-danger bp5-icon-cross"
					progressClassName="bp5-intent-danger"
					style={css.controlButton}
					hidden={!this.props.disk.action}
					disabled={this.state.disabled}
					onConfirm={(): void => {
						this.onAction('cancel',
							'Disk ' + actionLabel.toLowerCase() + ' canceled');
					}}
				/>
			</PageSave>
		</td>;
	}
}
