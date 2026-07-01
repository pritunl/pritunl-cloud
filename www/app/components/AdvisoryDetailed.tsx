/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from "@blueprintjs/core";
import * as Theme from '../Theme';
import * as MonacoEditor from "@monaco-editor/react";
import * as AdvisoryTypes from '../types/AdvisoryTypes';
import * as AdvisoryActions from '../actions/AdvisoryActions';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as MiscUtils from '../utils/MiscUtils';
import PageInfo from './PageInfo';
import * as PageInfos from './PageInfo';
import ConfirmButton from './ConfirmButton';
import CompletionStore from '../stores/CompletionStore';
import {severityClass, scoreLabel} from './Advisory';

const DESC_MAX_HEIGHT = 400;

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	advisory: AdvisoryTypes.AdvisoryRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	expanded: {[key: string]: boolean};
	expandedCves: boolean;
	expandedDismissedNodes: boolean;
	expandedDismissedInstances: boolean;
	descHeight: number;
	selected: {[key: string]: boolean};
	lastSelected: string;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 10px 10px',
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
	group: {
		flex: 1,
		minWidth: '280px',
		margin: '0 10px',
	} as React.CSSProperties,
	select: {
		margin: '7px 0px 0px 6px',
		paddingTop: '3px',
	} as React.CSSProperties,
	status: {
		margin: '6px 0 0 1px',
	} as React.CSSProperties,
	icon: {
		marginRight: '3px',
	} as React.CSSProperties,
	cards: {
		width: '100%',
		padding: '0 10px',
		marginTop: '4px',
	} as React.CSSProperties,
	descEditor: {
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
	count: {
		marginLeft: '6px',
		opacity: 0.7,
	} as React.CSSProperties,
	itemCard: {
		padding: '12px',
		marginBottom: '12px',
		borderLeftWidth: '4px',
		borderLeftStyle: 'solid',
	} as React.CSSProperties,
	itemHeader: {
		alignItems: 'center',
		gap: '8px',
		margin: '-12px -12px 10px -12px',
		padding: '10px 12px',
		background: 'rgba(138, 155, 168, 0.12)',
		borderBottom: '1px solid rgba(138, 155, 168, 0.25)',
		borderTopRightRadius: '3px',
	} as React.CSSProperties,
	headerTag: {
		paddingTop: '3px',
		paddingBottom: '3px',
	} as React.CSSProperties,
	title: {
		fontFamily: 'monospace',
		fontSize: '14px',
		fontWeight: 600,
	} as React.CSSProperties,
	tagRow: {
		marginBottom: '8px',
		gap: '6px',
	} as React.CSSProperties,
	tag: {
		paddingTop: '3px',
		paddingBottom: '3px',
		marginBottom: '4px',
	} as React.CSSProperties,
	description: {
		fontSize: '13px',
		whiteSpace: 'pre-wrap',
		wordBreak: 'break-word',
		padding: '6px 8px',
		background: 'rgba(138, 155, 168, 0.1)',
		borderRadius: '3px',
	} as React.CSSProperties,
	descriptionLimited: {
		display: '-webkit-box',
		WebkitLineClamp: 6,
		WebkitBoxOrient: 'vertical',
		overflow: 'hidden',
	} as React.CSSProperties,
	descriptionToggle: {
		marginTop: '4px',
		padding: '0',
		minHeight: '0',
		fontSize: '12px',
	} as React.CSSProperties,
	ipRow: {
		display: 'flex',
		alignItems: 'baseline',
		marginTop: '4px',
		gap: '8px',
	} as React.CSSProperties,
	ipLabel: {
		flex: '0 0 130px',
		fontSize: '12px',
		color: 'var(--bp5-text-color-muted, #5f6b7c)',
	} as React.CSSProperties,
	ipValue: {
		flex: '1',
		fontFamily: 'monospace',
		fontSize: '13px',
		wordBreak: 'break-all',
	} as React.CSSProperties,
	instCard: {
		padding: '5px 0 0 0',
		marginBottom: '8px',
	} as React.CSSProperties,
	checkBox: {
		display: 'flex',
		margin: '0 5px 0 5px',
	} as React.CSSProperties,
	check: {
		margin: "-5px -2px 0 -6px",
	} as React.CSSProperties,
	dismissSelected: {
		marginLeft: 'auto',
	} as React.CSSProperties,
	dismissedRow: {
		alignItems: 'center',
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
		margin: '0 5px',
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
};

export default class AdvisoryDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			expanded: {},
			expandedCves: false,
			expandedDismissedNodes: false,
			expandedDismissedInstances: false,
			descHeight: DESC_MAX_HEIGHT,
			selected: {},
			lastSelected: null,
		};
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		AdvisoryActions.remove(this.props.advisory.id).then((): void => {
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

	onDismiss = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		AdvisoryActions.dismiss(this.props.advisory.id).then((): void => {
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

	onRestore = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		AdvisoryActions.restore(this.props.advisory.id).then((): void => {
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

	onSelectResource = (resourceId: string, shift: boolean,
			orderedIds: string[]): void => {
		let selected = {
			...this.state.selected,
		};

		if (shift && this.state.lastSelected) {
			let start: number;
			let end: number;

			for (let i = 0; i < orderedIds.length; i++) {
				if (orderedIds[i] === resourceId) {
					start = i;
				} else if (orderedIds[i] === this.state.lastSelected) {
					end = i;
				}
			}

			if (start !== undefined && end !== undefined) {
				if (start > end) {
					end = [start, start = end][0];
				}

				for (let i = start; i <= end; i++) {
					selected[orderedIds[i]] = true;
				}

				this.setState({
					...this.state,
					selected: selected,
					lastSelected: resourceId,
				});

				return;
			}
		}

		if (selected[resourceId]) {
			delete selected[resourceId];
		} else {
			selected[resourceId] = true;
		}

		this.setState({
			...this.state,
			selected: selected,
			lastSelected: resourceId,
		});
	}

	onDismissSelected = (resourceIds: string[]): void => {
		if (!resourceIds.length) {
			return;
		}

		this.setState({
			...this.state,
			disabled: true,
		});
		AdvisoryActions.dismissResources(
				this.props.advisory.id, resourceIds).then((): void => {
			let selected = {
				...this.state.selected,
			};
			for (let resourceId of resourceIds) {
				delete selected[resourceId];
			}

			this.setState({
				...this.state,
				selected: selected,
			});

			setTimeout((): void => {
				this.setState({
					...this.state,
					disabled: false,
				});
			}, 1000);
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	onRestoreSelected = (resourceIds: string[]): void => {
		if (!resourceIds.length) {
			return;
		}

		this.setState({
			...this.state,
			disabled: true,
		});
		AdvisoryActions.restoreResources(
				this.props.advisory.id, resourceIds).then((): void => {
			let selected = {
				...this.state.selected,
			};
			for (let resourceId of resourceIds) {
				delete selected[resourceId];
			}

			this.setState({
				...this.state,
				selected: selected,
			});

			setTimeout((): void => {
				this.setState({
					...this.state,
					disabled: false,
				});
			}, 1000);
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	onDescMount = (editor: any): void => {
		let updateHeight = (): void => {
			let height = Math.min(DESC_MAX_HEIGHT, editor.getContentHeight());
			if (height !== this.state.descHeight) {
				this.setState({
					...this.state,
					descHeight: height,
				});
			}
		};
		editor.onDidContentSizeChange(updateHeight);
		updateHeight();
	}

	advisoryLink(advisoryRaw: string): string {
		let advisory = (advisoryRaw || "").replace(/[^a-zA-Z0-9:-]/g, '')
		if (advisory.startsWith('ALSA') || advisory.startsWith('RLSA') ||
				advisory.startsWith('RHSA')) {
			return `https://access.redhat.com/errata/RH${advisory.slice(2)}`
		} else if (advisory.startsWith('ELSA')) {
			return `https://linux.oracle.com/errata/${advisory}.html`
		} else if (advisory.startsWith('FEDORA')) {
			return `https://bodhi.fedoraproject.org/updates/${advisory}`
		}
		return ""
	}

	severityIntent(severity: string): Blueprint.Intent {
		switch ((severity || '').toLowerCase()) {
			case 'critical':
				return Blueprint.Intent.DANGER;
			case 'important':
			case 'high':
				return Blueprint.Intent.WARNING;
			case 'moderate':
			case 'medium':
				return Blueprint.Intent.PRIMARY;
			case 'low':
				return Blueprint.Intent.SUCCESS;
			default:
				return Blueprint.Intent.NONE;
		}
	}

	severityRank(severity: string): number {
		switch ((severity || '').toLowerCase()) {
			case 'critical':
				return 4;
			case 'important':
			case 'high':
				return 3;
			case 'moderate':
			case 'medium':
				return 2;
			case 'low':
				return 1;
			default:
				return 0;
		}
	}

	severityBarColor(severity: string): string {
		switch ((severity || '').toLowerCase()) {
			case 'critical':
				return 'var(--bp5-intent-danger, #cd4246)';
			case 'important':
			case 'high':
				return 'var(--bp5-intent-warning, #c87619)';
			case 'moderate':
			case 'medium':
				return 'var(--bp5-intent-primary, #215db0)';
			case 'low':
				return 'var(--bp5-intent-success, #1c6e42)';
			default:
				return 'rgba(138, 155, 168, 0.4)';
		}
	}

	stateIntent(state: string): Blueprint.Intent {
		switch ((state || '').toLowerCase()) {
			case 'running':
			case 'start':
				return Blueprint.Intent.SUCCESS;
			case 'starting':
			case 'provisioning':
				return Blueprint.Intent.PRIMARY;
			case 'destroy':
			case 'destroying':
				return Blueprint.Intent.DANGER;
			default:
				return Blueprint.Intent.NONE;
		}
	}

	renderDescription(key: string, text: string): JSX.Element {
		if (!text) {
			return null;
		}

		let expanded = !!this.state.expanded[key];
		let style = expanded ? css.description : {
			...css.description,
			...css.descriptionLimited,
		};

		return <React.Fragment>
			<div style={style}>
				{text}
			</div>
			<button
				className="bp5-button bp5-minimal bp5-small"
				type="button"
				style={css.descriptionToggle}
				onClick={(): void => {
					this.setState({
						...this.state,
						expanded: {
							...this.state.expanded,
							[key]: !expanded,
						},
					});
				}}
			>{expanded ? "Show less" : "Show more"}</button>
		</React.Fragment>;
	}

	renderIps(label: string, ips: string[]): JSX.Element {
		if (!ips || !ips.length) {
			return null;
		}

		return <div style={css.ipRow}>
			<span style={css.ipLabel}>{label}</span>
			<span style={css.ipValue}>{ips.join(', ')}</span>
		</div>;
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

	renderNodeCard(node: AdvisoryTypes.NodeInfo,
			orderedIds: string[], dismissed?: boolean): JSX.Element {
		let publicIps = node.public_ips && node.public_ips.length ?
			node.public_ips : ['-'];
		let publicIps6 = node.public_ips6 && node.public_ips6.length ?
			node.public_ips6 : ['-'];
		let privateIps = node.private_ips && node.private_ips.length ?
			node.private_ips : ['-'];

		let cardStyle = css.instCard;
		if (dismissed) {
			cardStyle = {
				...css.instCard,
				opacity: 0.5,
			};
		}

		return <Blueprint.Card
			key={node.id}
			compact={true}
			style={cardStyle}
		>
			<div className="layout horizontal flex" style={css.instBox}>
				<div className="layout center" style={css.checkBox}>
					<Blueprint.Checkbox
						style={css.check}
						alignIndicator={Blueprint.Alignment.RIGHT}
						checked={!!this.state.selected[node.id]}
						onClick={(evt): void => {
							this.onSelectResource(
								node.id, evt.shiftKey, orderedIds);
						}}
					/>
				</div>
				<div style={css.instItemFirst}>
					<PageInfo
						compact={true}
						style={css.instInfo}
						fields={[
							{
								label: 'Node',
								value: node.name || '-',
							},
							{
								label: 'Node ID',
								value: node.id || '-',
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
								label: 'Public IPv6',
								value: publicIps6,
								maxLines: 2,
							},
						]}
					/>
				</div>
			</div>
		</Blueprint.Card>;
	}

	renderInstanceCard(inst: AdvisoryTypes.InstanceInfo,
			orderedIds: string[], dismissed?: boolean): JSX.Element {
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

		let cardStyle = css.instCard;
		if (dismissed) {
			cardStyle = {
				...css.instCard,
				opacity: 0.5,
			};
		}

		return <Blueprint.Card
			key={inst.id}
			compact={true}
			style={cardStyle}
		>
			<div className="layout horizontal flex" style={css.instBox}>
				<div className="layout center" style={css.checkBox}>
					<Blueprint.Checkbox
						style={css.check}
						alignIndicator={Blueprint.Alignment.RIGHT}
						checked={!!this.state.selected[inst.id]}
						onClick={(evt): void => {
							this.onSelectResource(
								inst.id, evt.shiftKey, orderedIds);
						}}
					/>
				</div>
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
		</Blueprint.Card>;
	}

	renderVulnCard(vuln: AdvisoryTypes.Vulnerability): JSX.Element {
		let vulnId = vuln.id.split(":").pop();
		let nvdUrl = `https://access.redhat.com/security/cve/${vulnId}`;

		let tags: JSX.Element[] = [];
		if (vuln.vector === "network") {
			tags.push(<Blueprint.Tag key="vec" intent="danger"
				icon="globe-network" style={css.tag}>Network</Blueprint.Tag>);
		} else if (vuln.vector === "adjacent") {
			tags.push(<Blueprint.Tag key="vec" intent="warning"
				icon="globe-network" style={css.tag}>Adjacent</Blueprint.Tag>);
		} else if (vuln.vector === "local") {
			tags.push(<Blueprint.Tag key="vec" intent="success"
				icon="globe-network" style={css.tag}>Local</Blueprint.Tag>);
		} else if (vuln.vector === "physical") {
			tags.push(<Blueprint.Tag key="vec" intent="success"
				icon="globe-network" style={css.tag}>Physical</Blueprint.Tag>);
		}
		if (vuln.privileges === "none") {
			tags.push(<Blueprint.Tag key="priv" intent="danger"
				icon="shield" style={css.tag}>Unauthenticated</Blueprint.Tag>);
		} else if (vuln.privileges === "low") {
			tags.push(<Blueprint.Tag key="priv" intent="warning"
				icon="shield" style={css.tag}>Low Privileged</Blueprint.Tag>);
		} else if (vuln.privileges === "high") {
			tags.push(<Blueprint.Tag key="priv" intent="success"
				icon="shield" style={css.tag}>High Privileged</Blueprint.Tag>);
		}
		if (vuln.interaction === "none") {
			tags.push(<Blueprint.Tag key="int" intent="danger"
				icon="console" style={css.tag}>No Interaction</Blueprint.Tag>);
		} else if (vuln.interaction === "required") {
			tags.push(<Blueprint.Tag key="int" intent="success"
				icon="console" style={css.tag}>User Interaction</Blueprint.Tag>);
		}
		if (vuln.scope === "changed") {
			tags.push(<Blueprint.Tag key="scope" intent="warning"
				icon="route" style={css.tag}>Scope Changed</Blueprint.Tag>);
		}

		let sevIntent = this.severityIntent(vuln.severity || "");
		let sevLabel = MiscUtils.capitalize(vuln.severity || "Unknown");
		let scoreStr = vuln.score ? ` ${vuln.score.toFixed(1)}` : "";

		return <div key={vuln.id}
			className="bp5-card bp5-elevation-0"
			style={{
				...css.itemCard,
				borderLeftColor: this.severityBarColor(vuln.severity || ""),
			}}>
			<div className="layout horizontal" style={css.itemHeader}>
				<Blueprint.Tag intent={sevIntent} icon="shield"
					style={css.headerTag}>{sevLabel}{scoreStr}</Blueprint.Tag>
				<a
					href={nvdUrl}
					target="_blank"
					rel="noopener noreferrer"
					style={css.title}
				>{vulnId}</a>
			</div>
			{tags.length > 0 && <div className="layout horizontal wrap"
				style={css.tagRow}>
				{tags}
			</div>}
			{this.renderDescription(vuln.id, vuln.description)}
		</div>;
	}

	renderVulnerabilities(
			vulns: AdvisoryTypes.Vulnerability[]): JSX.Element {
		if (!vulns || !vulns.length) {
			return null;
		}

		let sorted = [...vulns].sort((a, b) => {
			let rank = this.severityRank(b.severity || "") -
				this.severityRank(a.severity || "");
			if (rank !== 0) {
				return rank;
			}
			return (b.score || 0) - (a.score || 0);
		});

		let important: AdvisoryTypes.Vulnerability[] = [];
		let other: AdvisoryTypes.Vulnerability[] = [];
		for (let vuln of sorted) {
			if (vuln.severity === "critical" || vuln.severity === "high") {
				important.push(vuln);
			} else {
				other.push(vuln);
			}
		}

		return <React.Fragment>
			<div style={css.section}>
				<span
					className="bp5-icon-standard bp5-icon-shield"
					style={css.sectionIcon}
				/>
				Vulnerabilities
			</div>
			{important.length > 0 ? <React.Fragment>
				<div style={css.section}>
					<span
						className="bp5-icon-standard bp5-icon-warning-sign bp5-text-intent-danger"
						style={css.sectionIcon}
					/>
					High Risk ({important.length})
				</div>
				{important.map((vuln): JSX.Element =>
					this.renderVulnCard(vuln))}
			</React.Fragment> : <div className="bp5-callout bp5-intent-success"
				style={{padding: '10px', marginBottom: '10px'}}>
				No high risk vulnerabilities.
			</div>}
			{other.length > 0 ? <React.Fragment>
				<button
					className={"bp5-button bp5-minimal " +
						(this.state.expandedCves ?
							"bp5-icon-chevron-down" :
							"bp5-icon-chevron-right")}
					type="button"
					style={{margin: '0 0 8px 0'}}
					onClick={(): void => {
						this.setState({
							...this.state,
							expandedCves: !this.state.expandedCves,
						});
					}}
				>
					Lower Risk ({other.length})
				</button>
				{this.state.expandedCves ? <div>
					{other.map((vuln): JSX.Element =>
						this.renderVulnCard(vuln))}
				</div> : null}
			</React.Fragment> : null}
		</React.Fragment>;
	}

	render(): JSX.Element {
		let advisory = this.props.advisory;

		let org = CompletionStore.organization(advisory.organization);

		let typeLabel = advisory.type;
		if (advisory.type === 'rhel') {
			typeLabel = 'Red Hat';
		}

		let statusClass = 'tab-close ' + severityClass(advisory.severity);

		let headerStyle = css.buttons;
		if (advisory.dismissed) {
			headerStyle = {
				...css.buttons,
				opacity: 0.5,
			};
		}

		let fields: PageInfos.Field[] = [
			{
				label: 'ID',
				value: advisory.id || 'Unknown',
				copy: true,
			},
			{
				label: 'Organization',
				value: org ? org.name : (advisory.organization || 'Global'),
			},
			{
				label: 'Reference',
				value: advisory.reference || '-',
				copy: true,
				link: this.advisoryLink(advisory.reference || ""),
			},
			{
				label: 'Type',
				value: typeLabel || '-',
			},
			{
				label: 'Severity',
				value: MiscUtils.capitalize(advisory.severity) || 'Unknown',
				valueClass: severityClass(advisory.severity),
			},
			{
				label: 'Score',
				value: scoreLabel(advisory.score),
			},
			{
				label: 'Updated',
				value: MiscUtils.formatDate(advisory.updated) || '-',
			},
		];

		let detailFields: PageInfos.Field[] = [];

		if (advisory.packages && advisory.packages.length) {
			detailFields.push({
				label: 'Packages',
				value: [...advisory.packages],
			});
		}

		let dismissals = new Set(advisory.dismissed_resources || []);

		let vulnerabilities = advisory.vulnerabilities || [];
		let nodes = (advisory.nodes_info || []).filter(
			(node): boolean => !dismissals.has(node.id));
		let instances = (advisory.instances_info || []).filter(
			(inst): boolean => !dismissals.has(inst.id));
		let dismissedNodes = (advisory.nodes_info || []).filter(
			(node): boolean => dismissals.has(node.id));
		let dismissedInstances = (advisory.instances_info || []).filter(
			(inst): boolean => dismissals.has(inst.id));

		let nodeIds = nodes.map((node): string => node.id);
		let instanceIds = instances.map((inst): string => inst.id);
		let dismissedNodeIds = dismissedNodes.map((node): string => node.id);
		let dismissedInstanceIds = dismissedInstances.map(
			(inst): string => inst.id);

		let selectedNodes = nodes.filter(
			(node): boolean => !!this.state.selected[node.id]).map(
			(node): string => node.id);
		let selectedInstances = instances.filter(
			(inst): boolean => !!this.state.selected[inst.id]).map(
			(inst): string => inst.id);
		let selectedDismissedNodes = dismissedNodes.filter(
			(node): boolean => !!this.state.selected[node.id]).map(
			(node): string => node.id);
		let selectedDismissedInstances = dismissedInstances.filter(
			(inst): boolean => !!this.state.selected[inst.id]).map(
			(inst): string => inst.id);

		return <td
			className="bp5-cell"
			colSpan={5}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div
					className="layout horizontal tab-close bp5-card-header"
					style={headerStyle}
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
							className="bp5-icon-standard bp5-icon-warning-sign"
						/>
						{MiscUtils.capitalize(advisory.severity) || 'Unknown'}
					</div>
					<div className="flex tab-close"/>
					<button
						className={"bp5-button bp5-minimal " +
							(advisory.dismissed ?
								"bp5-icon-undo" : "bp5-icon-disable")}
						style={css.button}
						type="button"
						disabled={this.state.disabled}
						onClick={advisory.dismissed ?
							this.onRestore : this.onDismiss}
					>{advisory.dismissed ? "Restore" : "Dismiss"}</button>
					<ConfirmButton
						className="bp5-minimal bp5-intent-danger bp5-icon-trash"
						style={css.button}
						safe={true}
						progressClassName="bp5-intent-danger"
						dialogClassName="bp5-intent-danger bp5-icon-delete"
						dialogLabel="Delete Advisory"
						confirmMsg="Permanently delete this advisory"
						confirmInput={true}
						items={[advisory.reference]}
						disabled={this.state.disabled}
						onConfirm={this.onDelete}
					/>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={fields}
					/>
				</div>
				<div style={css.group}>
					<PageInfo
						hidden={!detailFields.length}
						fields={detailFields}
					/>
				</div>
			</div>
			{advisory.description ? <div style={css.descEditor}>
				<div style={css.section}>
					<span
						className="bp5-icon-standard bp5-icon-align-left"
						style={css.sectionIcon}
					/>
					Description
				</div>
				<MonacoEditor.Editor
					height={this.state.descHeight + "px"}
					width="100%"
					defaultLanguage="markdown"
					theme={Theme.getEditorTheme()}
					value={advisory.description}
					onMount={this.onDescMount}
					options={{
						readOnly: true,
						folding: false,
						fontSize: 12,
						fontFamily: Theme.monospaceFont,
						fontWeight: Theme.monospaceWeight,
						tabSize: 4,
						detectIndentation: false,
						scrollBeyondLastLine: false,
						scrollbar: {
							alwaysConsumeMouseWheel: false,
						},
						minimap: {
							enabled: false,
						},
						wordWrap: "on",
						automaticLayout: true,
					}}
				/>
			</div> : null}
			<div style={css.cards}>
				{this.renderVulnerabilities(vulnerabilities)}
				{(nodes.length > 0 || dismissedNodes.length > 0) ?
						<React.Fragment>
					<div style={css.section}>
						<span
							className="bp5-icon-standard bp5-icon-cloud"
							style={css.sectionIcon}
						/>
						Nodes
						<span style={css.count}>({nodes.length})</span>
						<button
							className="bp5-button bp5-icon-disable"
							style={css.dismissSelected}
							type="button"
							disabled={this.state.disabled ||
								!selectedNodes.length}
							onClick={(): void => {
								this.onDismissSelected(selectedNodes);
							}}
						>Dismiss Selected</button>
					</div>
					{nodes.map((node): JSX.Element =>
						this.renderNodeCard(node, nodeIds))}
					{dismissedNodes.length > 0 ? <React.Fragment>
						<div className="layout horizontal"
							style={css.dismissedRow}>
							<button
								className={"bp5-button bp5-minimal " +
									(this.state.expandedDismissedNodes ?
										"bp5-icon-chevron-down" :
										"bp5-icon-chevron-right")}
								type="button"
								style={{margin: '0 0 8px 0'}}
								onClick={(): void => {
									this.setState({
										...this.state,
										expandedDismissedNodes:
											!this.state.expandedDismissedNodes,
									});
								}}
							>
								Dismissed ({dismissedNodes.length})
							</button>
							{this.state.expandedDismissedNodes ? <button
								className="bp5-button bp5-icon-undo"
								style={css.dismissSelected}
								type="button"
								disabled={this.state.disabled ||
									!selectedDismissedNodes.length}
								onClick={(): void => {
									this.onRestoreSelected(
										selectedDismissedNodes);
								}}
							>Restore Selected</button> : null}
						</div>
						{this.state.expandedDismissedNodes ? <div>
							{dismissedNodes.map((node): JSX.Element =>
								this.renderNodeCard(
									node, dismissedNodeIds, true))}
						</div> : null}
					</React.Fragment> : null}
				</React.Fragment> : null}
				{(instances.length > 0 || dismissedInstances.length > 0) ?
						<React.Fragment>
					<div style={css.section}>
						<span
							className="bp5-icon-standard bp5-icon-desktop"
							style={css.sectionIcon}
						/>
						Instances
						<span style={css.count}>({instances.length})</span>
						<button
							className="bp5-button bp5-icon-disable"
							style={css.dismissSelected}
							type="button"
							disabled={this.state.disabled ||
								!selectedInstances.length}
							onClick={(): void => {
								this.onDismissSelected(selectedInstances);
							}}
						>Dismiss Selected</button>
					</div>
					{instances.map((inst): JSX.Element =>
						this.renderInstanceCard(inst, instanceIds))}
					{dismissedInstances.length > 0 ? <React.Fragment>
						<div className="layout horizontal"
							style={css.dismissedRow}>
							<button
								className={"bp5-button bp5-minimal " +
									(this.state.expandedDismissedInstances ?
										"bp5-icon-chevron-down" :
										"bp5-icon-chevron-right")}
								type="button"
								style={{margin: '0 0 8px 0'}}
								onClick={(): void => {
									this.setState({
										...this.state,
										expandedDismissedInstances:
											!this.state.expandedDismissedInstances,
									});
								}}
							>
								Dismissed ({dismissedInstances.length})
							</button>
							{this.state.expandedDismissedInstances ? <button
								className="bp5-button bp5-icon-undo"
								style={css.dismissSelected}
								type="button"
								disabled={this.state.disabled ||
									!selectedDismissedInstances.length}
								onClick={(): void => {
									this.onRestoreSelected(
										selectedDismissedInstances);
								}}
							>Restore Selected</button> : null}
						</div>
						{this.state.expandedDismissedInstances ? <div>
							{dismissedInstances.map((inst): JSX.Element =>
								this.renderInstanceCard(
									inst, dismissedInstanceIds, true))}
						</div> : null}
					</React.Fragment> : null}
				</React.Fragment> : null}
			</div>
		</td>;
	}
}
