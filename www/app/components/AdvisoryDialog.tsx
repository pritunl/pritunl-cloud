/// <reference path="../References.d.ts"/>
import * as React from "react";
import * as Blueprint from "@blueprintjs/core";
import * as InstanceTypes from "../types/InstanceTypes";
import * as MiscUtils from "../utils/MiscUtils";

interface CveDetail {
	cve: string;
	detail: InstanceTypes.Advisory;
}

interface UpdateEntry {
	update: InstanceTypes.Update;
	cves: CveDetail[];
	importantCves: CveDetail[];
	worstScore: number;
	worstSeverity: string;
	link?: string;
}

interface State {
	open: boolean;
	showLowSeverity: boolean;
	expanded: {[key: string]: boolean};
	expandedStatements: {[key: string]: boolean};
	expandedCves: {[advisory: string]: boolean};
}

interface Props {
	updates: InstanceTypes.Update[];
}

const css = {
	dialog: {
		width: "90%",
		maxWidth: "720px",
	} as React.CSSProperties,
	body: {
		padding: "12px 16px 0 16px",
		maxHeight: "70vh",
		overflow: "auto",
	} as React.CSSProperties,
	header: {
		margin: "0 0 10px 0",
		fontWeight: 600,
	} as React.CSSProperties,
	count: {
		marginLeft: "6px",
		opacity: 0.7,
	} as React.CSSProperties,
	section: {
		display: "flex",
		alignItems: "center",
		margin: "14px 0 8px 0",
		fontWeight: 600,
	} as React.CSSProperties,
	updateCard: {
		padding: "12px",
		marginBottom: "12px",
		borderLeftWidth: "4px",
		borderLeftStyle: "solid",
	} as React.CSSProperties,
	updateHeader: {
		alignItems: "center",
		gap: "8px",
		margin: "-12px -12px 10px -12px",
		padding: "10px 12px",
		background: "rgba(138, 155, 168, 0.12)",
		borderBottom: "1px solid rgba(138, 155, 168, 0.25)",
		borderTopRightRadius: "3px",
	} as React.CSSProperties,
	cveCard: {
		padding: "10px",
		marginTop: "8px",
		background: "rgba(138, 155, 168, 0.06)",
		borderRadius: "3px",
	} as React.CSSProperties,
	headerRow: {
		alignItems: "center",
		marginBottom: "8px",
		gap: "8px",
	} as React.CSSProperties,
	headerTag: {
		paddingTop: "3px",
		paddingBottom: "3px",
	} as React.CSSProperties,
	title: {
		fontFamily: "monospace",
		fontSize: "14px",
		fontWeight: 600,
	} as React.CSSProperties,
	cveTitle: {
		fontFamily: "monospace",
		fontSize: "13px",
		fontWeight: 600,
	} as React.CSSProperties,
	tagRow: {
		marginBottom: "8px",
		gap: "6px",
	} as React.CSSProperties,
	tag: {
		paddingTop: "3px",
		paddingBottom: "3px",
		marginBottom: "4px",
	} as React.CSSProperties,
	packages: {
		fontSize: "13px",
		color: "var(--bp5-text-color-muted, #5f6b7c)",
		marginBottom: "6px",
		wordBreak: "break-all",
	} as React.CSSProperties,
	packageName: {
		fontFamily: "monospace",
		fontSize: "14px",
		fontWeight: 600,
		padding: "3px 8px",
		background: "rgba(138, 155, 168, 0.15)",
		borderRadius: "3px",
		wordBreak: "break-all",
	} as React.CSSProperties,
	packageList: {
		fontSize: "12px",
		fontFamily: "monospace",
		color: "var(--bp5-text-color-muted, #5f6b7c)",
		marginTop: "6px",
		padding: "6px 8px",
		background: "rgba(138, 155, 168, 0.08)",
		borderRadius: "3px",
		wordBreak: "break-all",
		whiteSpace: "pre-wrap",
	} as React.CSSProperties,
	description: {
		fontSize: "13px",
		whiteSpace: "pre-wrap",
		wordBreak: "break-word",
		padding: "6px 8px",
		background: "rgba(138, 155, 168, 0.1)",
		borderRadius: "3px",
	} as React.CSSProperties,
	descriptionLimited: {
		display: "-webkit-box",
		WebkitLineClamp: 6,
		WebkitBoxOrient: "vertical",
		overflow: "hidden",
	} as React.CSSProperties,
	statement: {
		fontSize: "13px",
		whiteSpace: "pre-wrap",
		wordBreak: "break-word",
		padding: "6px 8px",
		background: "rgba(138, 155, 168, 0.1)",
		borderRadius: "3px",
		marginTop: "6px",
	} as React.CSSProperties,
	statementLimited: {
		fontSize: "13px",
		whiteSpace: "pre-wrap",
		wordBreak: "break-word",
		padding: "6px 8px",
		background: "rgba(138, 155, 168, 0.1)",
		borderRadius: "3px",
		display: "-webkit-box",
		WebkitLineClamp: 6,
		WebkitBoxOrient: "vertical",
		overflow: "hidden",
		marginTop: "6px",
	} as React.CSSProperties,
	descriptionToggle: {
		marginTop: "4px",
		padding: "0",
		minHeight: "0",
		fontSize: "12px",
	} as React.CSSProperties,
	hiddenToggle: {
		marginTop: "8px",
		padding: "2px 6px",
		minHeight: "0",
		fontSize: "12px",
	} as React.CSSProperties,
}

export default class AdvisoryDialog extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			open: false,
			showLowSeverity: false,
			expanded: {},
			expandedStatements: {},
			expandedCves: {},
		}
	}

	rpmName(pkg: string): string {
		if (!pkg) {
			return pkg
		}
		let s = pkg
		let lastDot = s.lastIndexOf('.')
		if (lastDot > 0) {
			s = s.slice(0, lastDot)
		}
		let lastDash = s.lastIndexOf('-')
		if (lastDash > 0) {
			s = s.slice(0, lastDash)
		}
		lastDash = s.lastIndexOf('-')
		if (lastDash > 0) {
			s = s.slice(0, lastDash)
		}
		return s
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
		switch ((severity || "").toLowerCase()) {
			case "critical":
				return Blueprint.Intent.DANGER;
			case "high":
				return Blueprint.Intent.WARNING;
			case "medium":
				return Blueprint.Intent.PRIMARY;
			default:
				return Blueprint.Intent.NONE;
		}
	}

	severityBarColor(severity: string): string {
		switch ((severity || "").toLowerCase()) {
			case "critical":
				return "var(--bp5-intent-danger, #cd4246)";
			case "high":
				return "var(--bp5-intent-warning, #c87619)";
			case "medium":
				return "var(--bp5-intent-primary, #215db0)";
			default:
				return "rgba(138, 155, 168, 0.4)";
		}
	}

	isImportantCve(detail: InstanceTypes.Advisory): boolean {
		if (!detail) {
			return false;
		}
		if (detail.severity === "critical") {
			return true;
		}
		if (detail.vector === "network" &&
				(detail.severity === "high" || (detail.score || 0) >= 7)) {
			return true;
		}
		return false;
	}

	cveSortScore(detail: InstanceTypes.Advisory): number {
		let s = detail?.score || 0;
		if (detail?.vector === "network") s += 100;
		if (detail?.severity === "critical") s += 50;
		if (detail?.privileges === "none") s += 10;
		if (detail?.interaction === "none") s += 5;
		return s;
	}

	buildEntries(): UpdateEntry[] {
		let updates = this.props.updates;
		if (!updates) {
			return [];
		}

		let entries: UpdateEntry[] = [];
		for (let update of updates) {
			let cves = update.cves || [];
			let details = update.details || [];
			let detailMap: Record<string, InstanceTypes.Advisory> = {};
			for (let detail of details) {
				if (!detail || !detail.id) {
					continue;
				}
				let parts = detail.id.split(":");
				if (parts.length < 2) {
					continue;
				}
				detailMap[parts[1]] = detail;
			}
			let pairs: CveDetail[] = [];
			let seen = new Set<string>();
			for (let cve of cves) {
				if (!cve || seen.has(cve)) {
					continue;
				}
				let detail = detailMap[cve];
				if (!detail) {
					continue;
				}
				seen.add(cve);
				pairs.push({cve: cve, detail: detail});
			}

			pairs.sort((a, b) =>
				this.cveSortScore(b.detail) - this.cveSortScore(a.detail));

			let importantCves = pairs.filter(
				(p): boolean => this.isImportantCve(p.detail));

			let worstScore = 0;
			let worstSeverity = "";
			let severityRank: {[key: string]: number} = {
				"critical": 4,
				"high": 3,
				"medium": 2,
				"low": 1,
			};
			let worstRank = 0;
			for (let p of pairs) {
				let score = this.cveSortScore(p.detail);
				if (score > worstScore) {
					worstScore = score;
				}
				let sev = (p.detail.severity || "").toLowerCase();
				let rank = severityRank[sev] || 0;
				if (rank > worstRank) {
					worstRank = rank;
					worstSeverity = sev;
				}
			}

			entries.push({
				update: update,
				cves: pairs,
				importantCves: importantCves,
				worstScore: worstScore,
				worstSeverity: worstSeverity,
				link: this.advisoryLink(update.advisory || ""),
			});
		}

		return entries;
	}

	buttonIntent(entries: UpdateEntry[]): string {
		if (entries.length === 0) {
			return "";
		}

		let hasHigh = false;
		for (let entry of entries) {
			if (entry.worstSeverity === "critical") {
				return "bp5-intent-danger";
			}
			if (entry.worstSeverity === "high") {
				hasHigh = true;
			}
		}

		return hasHigh ? "bp5-intent-warning" : "bp5-intent-primary";
	}

	renderCveCard(entry: UpdateEntry, pair: CveDetail): JSX.Element {
		let d = pair.detail;
		let key = (entry.update.advisory || "") + "|" + pair.cve;
		let nvdUrl = `https://access.redhat.com/security/cve/${pair.cve}`;

		let tags: JSX.Element[] = [];
		if (d.vector === "network") {
			tags.push(<Blueprint.Tag key="vec"
				intent="danger"
				icon="globe-network"
				style={css.tag}>Network</Blueprint.Tag>);
		} else if (d.vector === "adjacent") {
			tags.push(<Blueprint.Tag key="vec"
				intent="warning"
				icon="globe-network"
				style={css.tag}>Adjacent</Blueprint.Tag>);
		} else if (d.vector === "local") {
			tags.push(<Blueprint.Tag key="vec"
				intent="success"
				icon="globe-network"
				style={css.tag}>Local</Blueprint.Tag>);
		} else if (d.vector === "physical") {
			tags.push(<Blueprint.Tag key="vec"
				intent="success"
				icon="globe-network"
				style={css.tag}>Physical</Blueprint.Tag>);
		}
		if (d.privileges === "none") {
			tags.push(<Blueprint.Tag key="priv"
				intent="danger"
				icon="shield"
				style={css.tag}>Unauthenticated</Blueprint.Tag>);
		} else if (d.privileges === "low") {
			tags.push(<Blueprint.Tag key="priv"
				intent="warning"
				icon="shield"
				style={css.tag}>Low Privileged</Blueprint.Tag>);
		} else if (d.privileges === "high") {
			tags.push(<Blueprint.Tag key="priv"
				intent="success"
				icon="shield"
				style={css.tag}>High Privileged</Blueprint.Tag>);
		}
		if (d.interaction === "none") {
			tags.push(<Blueprint.Tag key="int"
				intent="danger"
				icon="console"
				style={css.tag}>No Interaction</Blueprint.Tag>);
		} else if (d.interaction === "required") {
			tags.push(<Blueprint.Tag key="int"
				intent="success"
				icon="console"
				style={css.tag}>User Interaction</Blueprint.Tag>);
		}
		if (d.scope === "changed") {
			tags.push(<Blueprint.Tag key="scope"
				intent="warning"
				icon="route"
				style={css.tag}>Scope Changed</Blueprint.Tag>);
		}

		let sevIntent = this.severityIntent(d.severity || "");
		let sevLabel = MiscUtils.capitalize(d.severity || "Unknown");
		let scoreLabel = d.score ? ` ${d.score.toFixed(1)}` : "";

		return <div key={pair.cve} style={css.cveCard}>
			<div className="layout horizontal" style={css.headerRow}>
				<Blueprint.Tag
					intent={sevIntent}
					icon="shield"
					style={css.headerTag}>{sevLabel}{scoreLabel}</Blueprint.Tag>
				<a
					href={nvdUrl}
					target="_blank"
					rel="noopener noreferrer"
					style={css.cveTitle}
				>{pair.cve}</a>
			</div>
			{tags.length > 0 && <div className="layout horizontal wrap"
				style={css.tagRow}>
				{tags}
			</div>}
			{this.renderDescription(key, d.description, d.statement)}
		</div>;
	}

	renderDescription(key: string, text: string, statement: string): JSX.Element {
		let expanded = !!this.state.expanded[key];
		let style = expanded ? css.description : {
			...css.description,
			...css.descriptionLimited,
		};
		let statementStyle = expanded ? css.statement : css.statementLimited;

		return <>
			<div style={style}>
				{text}
			</div>
			<div style={statementStyle} hidden={!statement}>
				{statement}
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
		</>;
	}

	renderStatement(key: string, text: string): JSX.Element {
		let expanded = !!this.state.expandedStatements[key];
		let style = expanded ? css.statement : css.statementLimited;

		return <>
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
						expandedStatements: {
							...this.state.expanded,
							[key]: !expanded,
						},
					});
				}}
			>{expanded ? "Show less" : "Show more"}</button>
		</>;
	}

	renderUpdateCard(entry: UpdateEntry): JSX.Element {
		let update = entry.update;
		let sevIntent = this.severityIntent(entry.worstSeverity);
		let sevLabel = entry.worstSeverity ?
			MiscUtils.capitalize(entry.worstSeverity) : "Unknown";

		let advisoryKey = update.advisory || "";
		let cvesExpanded = !!this.state.expandedCves[advisoryKey];
		let cvesToShow = cvesExpanded ? entry.cves : entry.importantCves;
		let hiddenCount = entry.cves.length - entry.importantCves.length;

		let packages = update.packages || [];
		let primaryName = packages.length > 0 ? this.rpmName(packages[0]) : "";
		let hasFullVersionInfo = packages.length > 1 ||
			(packages.length === 1 && packages[0] !== primaryName);

		let detailsKey = advisoryKey + "|details";
		let detailsExpanded = !!this.state.expanded[detailsKey];
		let hasDescription = !!update.description;
		let showDetailsToggle = hasDescription || hasFullVersionInfo;
		let descriptionStyle = detailsExpanded ? css.description : {
			...css.description,
			...css.descriptionLimited,
		};

		return <div key={update.advisory}
			className="bp5-card bp5-elevation-0"
			style={{
				...css.updateCard,
				borderLeftColor: this.severityBarColor(entry.worstSeverity),
			}}>
			<div className="layout horizontal" style={css.updateHeader}>
				<Blueprint.Tag
					large={true}
					intent={sevIntent}
					icon="shield"
					style={css.headerTag}>{sevLabel}</Blueprint.Tag>
				{primaryName && <span style={css.packageName}>
					{primaryName}
				</span>}
				{entry.link ? <a
					href={entry.link}
					target="_blank"
					rel="noopener noreferrer"
					style={css.title}
				>{update.advisory}</a> : <span style={css.title}>
					{update.advisory}
				</span>}
			</div>
			{hasDescription && <div style={descriptionStyle}>
				{update.description}
			</div>}
			{detailsExpanded && hasFullVersionInfo &&
				<div style={css.packageList}>
					{packages.join("\n")}
				</div>}
			{showDetailsToggle && <button
				className="bp5-button bp5-minimal bp5-small"
				type="button"
				style={css.descriptionToggle}
				onClick={(): void => {
					this.setState({
						...this.state,
						expanded: {
							...this.state.expanded,
							[detailsKey]: !detailsExpanded,
						},
					});
				}}
			>{detailsExpanded ? "Show less" : "Show more"}</button>}
			{cvesToShow.map((p): JSX.Element =>
				this.renderCveCard(entry, p))}
			<div>
				{hiddenCount > 0 && <button
					className={"bp5-button bp5-minimal bp5-small " +
						(cvesExpanded ?
							"bp5-icon-chevron-up" :
							"bp5-icon-chevron-down")}
					type="button"
					style={css.hiddenToggle}
					onClick={(): void => {
						this.setState({
							...this.state,
							expandedCves: {
								...this.state.expandedCves,
								[advisoryKey]: !cvesExpanded,
							},
						});
					}}
				>{cvesExpanded ?
					`Hide ${hiddenCount} additional CVE${hiddenCount === 1 ? "" : "s"}` :
					`Show ${hiddenCount} additional CVE${hiddenCount === 1 ? "" : "s"}`}</button>}
			</div>
		</div>;
	}

	renderBody(entries: UpdateEntry[]): JSX.Element {
		if (entries.length === 0) {
			return <div style={css.body}>
				<div className="bp5-callout bp5-intent-success"
					style={{padding: "12px"}}
				>
					No outstanding security advisories reported by the guest agent.
				</div>
			</div>;
		}

		entries.sort((a, b) => b.worstScore - a.worstScore);

		let important: UpdateEntry[] = [];
		let other: UpdateEntry[] = [];
		for (let entry of entries) {
			if (entry.importantCves.length > 0) {
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
						style={{marginRight: "6px"}}
					/>
					High Risk ({important.length})
				</div>
				{important.map((e): JSX.Element => this.renderUpdateCard(e))}
			</> : <div className="bp5-callout bp5-intent-success"
				style={{padding: "10px", marginBottom: "10px"}}>
				No remotely exploitable or critical advisories.
			</div>}
			{other.length > 0 ? <>
				<button
					className={"bp5-button bp5-minimal " +
						(this.state.showLowSeverity ?
							"bp5-icon-chevron-down" :
							"bp5-icon-chevron-right")}
					type="button"
					style={{margin: "8px 0"}}
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
					{other.map((e): JSX.Element => this.renderUpdateCard(e))}
				</div> : null}
			</> : null}
		</div>;
	}

	render(): JSX.Element {
		let entries = this.buildEntries();

		let dialog: JSX.Element
		if (this.state.open) {
			dialog = <Blueprint.Dialog
				title={
					<div>
						Security Advisories
						<span style={css.count}>
							({entries.length} advisor{entries.length === 1 ? "y" : "ies"})
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
					this.buttonIntent(entries)}
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
