/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as InstanceTypes from '../types/InstanceTypes';
import * as MiscUtils from '../utils/MiscUtils';

interface CveEntry {
	cve: string;
	detail: InstanceTypes.Advisory;
	packages: string[];
	advisories: string[];
}

interface State {
	open: boolean;
	showLowSeverity: boolean;
}

interface Props {
	instance: InstanceTypes.InstanceRo;
}

const css = {
	dialog: {
		width: '90%',
		maxWidth: '720px',
	} as React.CSSProperties,
	body: {
		padding: '12px 16px',
		maxHeight: '70vh',
		overflow: 'auto',
	} as React.CSSProperties,
	header: {
		margin: '0 0 10px 0',
		fontWeight: 600,
	} as React.CSSProperties,
	count: {
		marginLeft: '6px',
		opacity: 0.7,
	} as React.CSSProperties,
	section: {
		display: 'flex',
		alignItems: 'center',
		margin: '14px 0 8px 0',
		fontWeight: 600,
	} as React.CSSProperties,
	card: {
		padding: '12px',
		marginBottom: '10px',
	} as React.CSSProperties,
	headerRow: {
		alignItems: 'center',
		marginBottom: '8px',
		gap: '8px',
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
		marginRight: '6px',
		marginBottom: '4px',
	} as React.CSSProperties,
	packages: {
		fontSize: '11px',
		color: 'var(--bp5-text-color-muted, #5f6b7c)',
		marginBottom: '6px',
		wordBreak: 'break-all',
	} as React.CSSProperties,
	description: {
		fontSize: '12px',
		whiteSpace: 'pre-wrap',
		wordBreak: 'break-word',
		maxHeight: '160px',
		overflow: 'auto',
		padding: '6px 8px',
		background: 'rgba(138, 155, 168, 0.1)',
		borderRadius: '3px',
	} as React.CSSProperties,
}

export default class InstanceAdvDialog extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			open: false,
			showLowSeverity: false,
		}
	}

	buttonIntent(): string {
		let entries = this.aggregateCves();
		if (entries.length === 0) {
			return '';
		}

		let hasHigh = false;
		for (let entry of entries) {
			let sev = (entry.detail?.severity || '').toLowerCase();
			if (sev === 'critical') {
				return 'bp5-intent-danger';
			}
			if (sev === 'high') {
				hasHigh = true;
			}
		}

		return hasHigh ? 'bp5-intent-warning' : 'bp5-intent-primary';
	}

	severityIntent(severity: string): string {
		switch ((severity || '').toLowerCase()) {
			case 'critical':
				return 'danger';
			case 'high':
				return 'warning';
			case 'medium':
				return 'primary';
			default:
				return 'none';
		}
	}

	aggregateCves(): CveEntry[] {
		let updates = this.props.instance.guest?.updates;

		if (!updates) {
			return [];
		}

		let map = new Map<string, CveEntry>();
		for (let update of updates) {
			let cves = update.cves || [];
			let details = update.details || [];
			for (let i = 0; i < cves.length; i++) {
				let cve = cves[i];
				let detail = details[i];
				if (!cve || !detail) {
					continue;
				}
				let entry = map.get(cve);
				if (!entry) {
					entry = {
						cve: cve,
						detail: detail,
						packages: [],
						advisories: [],
					};
					map.set(cve, entry);
				}
				if (update.package &&
						entry.packages.indexOf(update.package) === -1) {
					entry.packages.push(update.package);
				}
				for (let adv of (update.advisories || [])) {
					if (entry.advisories.indexOf(adv) === -1) {
						entry.advisories.push(adv);
					}
				}
			}
		}

		return Array.from(map.values());
	}

	isImportantCve(entry: CveEntry): boolean {
		let d = entry.detail;
		if (!d) {
			return false;
		}
		if (d.severity === 'critical') {
			return true;
		}
		if (d.vector === 'network' &&
				(d.severity === 'high' || (d.score || 0) >= 7)) {
			return true;
		}
		return false;
	}

	cveSortScore(entry: CveEntry): number {
		let d = entry.detail;
		let s = d?.score || 0;
		if (d?.vector === 'network') s += 100;
		if (d?.severity === 'critical') s += 50;
		if (d?.privileges === 'none') s += 10;
		if (d?.interaction === 'none') s += 5;
		return s;
	}

	renderCveCard(entry: CveEntry): JSX.Element {
		let d = entry.detail;
		let nvdUrl = `https://nvd.nist.gov/vuln/detail/${entry.cve}`;

		let tags: JSX.Element[] = [];
		if (d.vector === 'network') {
			tags.push(<Blueprint.Tag key="vec"
				intent="danger"
				icon="globe-network"
				style={css.tag}>Network</Blueprint.Tag>);
		} else if (d.vector === 'adjacent') {
			tags.push(<Blueprint.Tag key="vec"
				intent="warning"
				icon="globe-network"
				style={css.tag}>Adjacent</Blueprint.Tag>);
		} else if (d.vector === 'local') {
			tags.push(<Blueprint.Tag key="vec"
				intent="success"
				icon="globe-network"
				style={css.tag}>Local</Blueprint.Tag>);
		} else if (d.vector === 'physical') {
			tags.push(<Blueprint.Tag key="vec"
				intent="success"
				icon="globe-network"
				style={css.tag}>Physical</Blueprint.Tag>);
		}
		if (d.privileges === 'none') {
			tags.push(<Blueprint.Tag key="priv"
				intent="danger"
				icon="shield"
				style={css.tag}>Unauthenticated</Blueprint.Tag>);
		} else if (d.privileges === 'low') {
			tags.push(<Blueprint.Tag key="priv"
				intent="warning"
				icon="shield"
				style={css.tag}>Low Privileged</Blueprint.Tag>);
		} else if (d.privileges === 'high') {
			tags.push(<Blueprint.Tag key="priv"
				intent="success"
				icon="shield"
				style={css.tag}>High Privileged</Blueprint.Tag>);
		}
		if (d.interaction === 'none') {
			tags.push(<Blueprint.Tag key="int"
				intent="danger"
				icon="console"
				style={css.tag}>No Interaction</Blueprint.Tag>);
		} else if (d.interaction === 'required') {
			tags.push(<Blueprint.Tag key="int"
				intent="success"
				icon="console"
				style={css.tag}>User Interaction</Blueprint.Tag>);
		}
		if (d.complexity === 'low' && d.vector === 'network') {
			tags.push(<Blueprint.Tag key="cplx"
				intent="warning"
				style={css.tag}>Easy to exploit</Blueprint.Tag>);
		}
		if (d.scope === 'changed') {
			tags.push(<Blueprint.Tag key="scope"
				intent="warning"
				icon="route"
				style={css.tag}>Scope changed</Blueprint.Tag>);
		}

		let sevIntent = this.severityIntent(d.severity || '');
		let sevLabel = MiscUtils.capitalize(d.severity || 'Unknown');
		let scoreLabel = d.score ? ` ${d.score.toFixed(1)}` : '';

		return <div key={entry.cve}
			className="bp5-card bp5-elevation-0"
			style={css.card}>
			<div className="layout horizontal" style={css.headerRow}>
				<span
					className={`bp5-tag bp5-intent-${sevIntent} bp5-large`}
					style={css.tag}
				>{sevLabel}{scoreLabel}</span>
				<a
					href={nvdUrl}
					target="_blank"
					rel="noopener noreferrer"
					style={css.title}
				>{entry.cve}</a>
			</div>
			{tags.length > 0 && <div className="layout horizontal wrap"
				style={css.tagRow}>
				{tags}
			</div>}
			{entry.packages.length > 0 && <div style={css.packages}>
				{entry.packages.length === 1 ? 'Package: ' :
					`Packages (${entry.packages.length}): `}
				{entry.packages.join(', ')}
			</div>}
			{d.description && <div style={css.description}>
				{d.description}
			</div>}
		</div>;
	}

	renderBody(entries: CveEntry[]): JSX.Element {
		if (entries.length === 0) {
			return <div style={css.body}>
				<div className="bp5-callout bp5-intent-success"
					style={{padding: '12px'}}>
					<h5 className="bp5-heading">No security advisories</h5>
					No outstanding security advisories reported by the guest agent.
				</div>
			</div>;
		}

		entries.sort((a, b) =>
			this.cveSortScore(b) - this.cveSortScore(a));

		let important: CveEntry[] = [];
		let other: CveEntry[] = [];
		for (let entry of entries) {
			if (this.isImportantCve(entry)) {
				important.push(entry);
			} else {
				other.push(entry);
			}
		}

		return <div style={css.body}>
			{important.length > 0 ? <>
				<div style={css.section}>
					<span
						className="bp5-icon-standard bp5-icon-warning-sign bp5-text-intent-danger"
						style={{marginRight: '6px'}}
					/>
					High Risk ({important.length})
				</div>
				{important.map((e): JSX.Element => this.renderCveCard(e))}
			</> : <div className="bp5-callout bp5-intent-success"
				style={{padding: '10px', marginBottom: '10px'}}>
				No remotely exploitable or critical advisories.
			</div>}
			{other.length > 0 ? <>
				<button
					className={"bp5-button bp5-minimal " +
						(this.state.showLowSeverity ?
							"bp5-icon-chevron-down" :
							"bp5-icon-chevron-right")}
					type="button"
					style={{margin: '8px 0'}}
					onClick={(): void => {
						this.setState({
							...this.state,
							showLowSeverity: !this.state.showLowSeverity,
						});
					}}
				>
					Lower Risk ({other.length})
				</button>
				{this.state.showLowSeverity ? <div>
					{other.map((e): JSX.Element => this.renderCveCard(e))}
				</div> : null}
			</> : null}
		</div>;
	}

	render(): JSX.Element {
		let dialog: JSX.Element
		if (this.state.open) {
			let entries = this.aggregateCves();

			dialog = <Blueprint.Dialog
				title={
					<div>
						Security Advisories
						<span style={css.count}>
							({entries.length} CVE{entries.length === 1 ? '' : 's'})
						</span>
					</div>
				}
				style={css.dialog}
				isOpen={this.state.open}
				usePortal={true}
				portalContainer={document.body}
				onClose={(): void => {
					this.setState({
						...this.state,
						open: false,
					})
				}}
			>
				{this.renderBody(entries)}
				<div className="bp5-dialog-footer">
					<div className="bp5-dialog-footer-actions">
						<button
							className="bp5-button bp5-intent-danger"
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									open: false,
								})
							}}
						>Close</button>
					</div>
				</div>
			</Blueprint.Dialog>
		}

		return <div>
			<button
				className={"bp5-button bp5-minimal bp5-icon-shield " +
					this.buttonIntent()}
				type="button"
				onClick={(): void => {
					this.setState({
						...this.state,
						open: true,
					})
				}}
			>Security Advisories</button>
			{dialog}
		</div>
	}
}
