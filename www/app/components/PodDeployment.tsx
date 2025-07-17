/// <reference path="../References.d.ts"/>
import * as React from "react"
import * as Blueprint from "@blueprintjs/core"
import * as Theme from '../Theme';
import * as PodTypes from "../types/PodTypes"
import * as PodActions from "../actions/PodActions"
import * as InstanceActions from '../actions/InstanceActions';
import * as MiscUtils from '../utils/MiscUtils';
import PodDeploymentEdit from "./PodDeploymentEdit";
import PageInfo from "./PageInfo"
import Editor from "./Editor"
import * as Router from "../Router";
import * as PageInfos from './PageInfo';

interface Props {
	hidden: boolean
	disabled: boolean
	selected: boolean
	commitMap: Record<string, PodTypes.Commit>
	deployment: PodTypes.Deployment
	onSelect: (shift: boolean) => void
}

interface State {
	logsOpen: boolean
	editOpen: boolean
}

const css = {
	container: {
		height: "900px",
		overflowY: "auto",
		marginBottom: "10px",
	} as React.CSSProperties,
	box: {
		fontSize: "11px",
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
	boxEmpty: {
		flex: 1,
		margin: "20px 10px 20px 10px",
	} as React.CSSProperties,
	card: {
		padding: "5px 5px 3px 5px",
	} as React.CSSProperties,
	cardInactive: {
		padding: "5px 5px 3px 5px",
		opacity: 0.6,
	} as React.CSSProperties,
	check: {
		margin: "0 0 0 0",
	} as React.CSSProperties,
	checkBox: {
		display: "flex",
		paddingBottom: "2px",
	} as React.CSSProperties,
	info: {
		marginBottom: "0px",
		fontSize: "11px",
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
	itemFirst: {
		flex: "1 1 auto",
		minWidth: "100px",
		maxWidth: "170px",
		margin: "0 5px 0 -4px",
	} as React.CSSProperties,
	item: {
		flex: "1 1 auto",
		width: 0,
		margin: " 0 5px",
	} as React.CSSProperties,
	itemSmall: {
		flex: "0.7 1 auto",
		width: 0,
		margin: " 0 5px",
	} as React.CSSProperties,
	itemMedium: {
		flex: "0.8 1 auto",
		width: 0,
		margin: " 0 5px",
	} as React.CSSProperties,
	itemFull: {
		flex: "2 1 auto",
		width: 0,
		margin: " 0 5px",
	} as React.CSSProperties,
	itemLast: {
		flex: "0 1 auto",
		minWidth: "123px",
		margin: " 0 5px",
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	hoverInfo: {
		padding: "10px",
		fontSize: "12px",
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
	cardButton: {
		marginTop: "1px",
		marginRight: "5px",
	} as React.CSSProperties,
	cardButtonRight: {
		marginTop: "6px",
	} as React.CSSProperties,
	editor: {
		marginTop: "5px",
	} as React.CSSProperties,
}

export default class PodDeployment extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context)
		this.state = {
			logsOpen: false,
			editOpen: false,
		}
	}

	onLogsToggle = (): void => {
		this.setState({
			...this.state,
			logsOpen: !this.state.logsOpen,
		})
	}

	onEditToggle = (): void => {
		this.setState({
			...this.state,
			editOpen: !this.state.editOpen,
		})
	}

	onEditClose = (): void => {
		this.setState({
			...this.state,
			editOpen: false,
		})
	}

	render(): JSX.Element {
		if (this.props.hidden) {
			return <div></div>
		}

		let deployment = this.props.deployment

		let label = "deployment"
		let labelTitle = "Deployment"
		if (deployment.kind == "image") {
			label = "image"
			labelTitle = "Image"
		}

		let cardStyle = css.card
		if (deployment.state === "archived") {
			cardStyle = css.cardInactive
		}

		let stateValue = MiscUtils.capitalize(deployment.state) || "-"
		let stateClass = ""
		switch (deployment.state) {
			case "deployed":
				stateClass = "bp5-text-intent-success"
				break
			case "reserved":
				stateClass = "bp5-text-intent-primary"
				break
			case "archived":
				stateClass = "bp5-text-intent-warning"
				break
		}

		switch (deployment.action) {
			case "migrate":
				stateValue = "Migrating"
				stateClass = "bp5-text-intent-warning"
				break
			case "restore":
				stateValue = "Restoring"
				stateClass = "bp5-text-intent-primary"
				break
			case "destroy":
				stateValue = "Destroying"
				stateClass = "bp5-text-intent-danger"
				break
			case "archive":
				stateValue = "Archiving"
				stateClass = "bp5-text-intent-warning"
				break
		}

		let statusClass = ""
		switch (deployment.instance_status) {
			case "Running":
				statusClass = "bp5-text-intent-success"
				break
			case "Starting":
			case "Restarting":
			case "Updating":
			case "Provisioning":
				statusClass = "bp5-text-intent-primary"
				break
			case "Failed":
			case "Stopping":
			case "Destroying":
			case "Stopped":
				statusClass = "bp5-text-intent-danger"
				break
		}

		let instData = deployment.instance_data || {}

		let publicIps = instData.public_ips
		if (!publicIps || !publicIps.length) {
			publicIps = ["-"]
		}

		let publicIps6 = instData.public_ips6
		if (!publicIps6 || !publicIps6.length) {
			publicIps6 = ["-"]
		}

		let privateIps = instData.private_ips
		if (!privateIps || !privateIps.length) {
			privateIps = ["-"]
		}

		let privateIps6 = instData.private_ips6
		if (!privateIps6 || !privateIps6.length) {
			privateIps6 = ["-"]
		}

		let commitClass = deployment.spec_offset === 0 ?
			"bp5-text-intent-success" :	"bp5-text-intent-danger"

		let specHover = <div
			className="bp5-content-popover"
			style={css.hoverInfo}
		>
			<PageInfo
				compact={true}
				style={css.info}
				fields={[
					{
						label: "Commit",
						value: deployment.spec.substring(12) || "-",
					},
					{
						label: "Timestamp",
						value: MiscUtils.formatDateLocal(deployment.spec_timestamp),
					},
					{
						label: "Behind",
						value: deployment.spec_offset,
						valueClass: commitClass,
					},
				]}
			/>
		</div>

		let domainInfo: PageInfos.Field
		if (deployment?.domain_data?.records) {
			let domainFields: PageInfos.Field[] = [];

			for (let rec of deployment.domain_data.records) {
				domainFields.push({
					label: rec.domain,
					value: rec.value,
				})
			}

			let domainHover = <div
				className="bp5-content-popover"
				style={css.hoverInfo}
			>
				<PageInfo
					compact={true}
					style={css.info}
					fields={domainFields}
				/>
			</div>

			domainInfo = {
				label: "Domains",
				value: "Registered",
				valueClass: "bp5-text-intent-success",
				hover: domainHover,
			}
		}

		let deplyStatus = MiscUtils.capitalize(deployment.status) || "-"
		let heartbeatClass = ""
		if (deployment.status === "healthy") {
			heartbeatClass = "bp5-text-intent-success"
		} else {
			heartbeatClass = "bp5-text-intent-danger"
		}

		let agentStatus = MiscUtils.capitalize(
			deployment.instance_guest_status) || "-"
		let agentClass = heartbeatClass
		switch (deployment.instance_guest_status) {
			case "initializing":
				agentClass = "bp5-text-intent-primary"
				break
			case "reloading_clean":
				agentStatus = "Reloading"
				agentClass = "bp5-text-intent-primary"
				break
			case "reloading_fault":
				agentStatus = "Reloading"
				agentClass = "bp5-text-intent-danger"
				break
			case "fault":
				agentClass = "bp5-text-intent-danger"
				break
		}

		let heartbeatHover = <div
			className="bp5-content-popover"
			style={css.hoverInfo}
		>
			<PageInfo
				compact={true}
				style={css.info}
				fields={[
					{
						label: "Status",
						value: deplyStatus,
						valueClass: heartbeatClass,
					},
					{
						label: "Heartbeat Timestamp",
						value: MiscUtils.formatDateLocal(
							deployment.instance_heartbeat) || "-",
						valueClass: heartbeatClass,
					},
				]}
			/>
		</div>

		let resourceBars: PageInfos.Bar[] = []
		resourceBars.push({
			label: "Instance Resources",
			progressClass: 'bp5-no-stripes bp5-intent-success',
			value: deployment.instance_load1 || 0,
		})
		resourceBars.push({
			progressClass: 'bp5-no-stripes bp5-intent-warning',
			value: deployment.instance_load5 || 0,
		})
		resourceBars.push({
			progressClass: 'bp5-no-stripes bp5-intent-danger',
			value: deployment.instance_load15 || 0,
		})
		resourceBars.push({
			progressClass: 'bp5-no-stripes bp5-intent-primary',
			value: deployment.instance_memory_usage || 0,
		})

		let editor: JSX.Element
		if (this.state.logsOpen) {
			editor = <Editor
				height="500px"
				interval={1000}
				style={css.editor}
				autoScroll={true}
				readOnly={true}
				refresh={async (first: boolean): Promise<string> => {
					try {
						let logs = await PodActions.log(
							this.props.deployment, "agent", !first)
						return logs.join("")
					} catch (error) {
						return ""
					}
				}}
			/>
		}

		if (deployment.kind === "image" && deployment.image_id) {
			return <Blueprint.Card
				key={deployment.id}
				compact={true}
				style={cardStyle}
			>
				<div className="layout vertical flex">
					<div className="layout horizontal flex" style={css.box}>
						<div className="layout center" style={css.checkBox}>
							<Blueprint.Checkbox
								style={css.check}
								checked={this.props.selected}
								onClick={(evt): void => {
									this.props.onSelect(evt.shiftKey)
								}}
							/>
						</div>
						<div style={css.itemFirst}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: "Build ID",
										value: deployment.id,
									},
								]}
							/>
							<button
								className="bp5-button bp5-small"
								style={css.cardButton}
								hidden={this.state.logsOpen}
								onClick={this.onLogsToggle}
							>Logs</button>
							<button
								className="bp5-button bp5-small bp5-active bp5-intent-danger"
								style={css.cardButton}
								hidden={!this.state.logsOpen}
								onClick={this.onLogsToggle}
							>Logs</button>
							<button
								className="bp5-button bp5-small"
								style={css.cardButton}
								hidden={this.state.editOpen}
								onClick={this.onEditToggle}
							>Settings</button>
							<button
								className="bp5-button bp5-small bp5-active bp5-intent-danger"
								style={css.cardButton}
								hidden={!this.state.editOpen}
								onClick={this.onEditToggle}
							>Settings</button>
						</div>
						<div style={css.item}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: "Commit ID",
										value: deployment.spec.substring(12),
										hover: specHover,
										valueClass: commitClass,
									},
								]}
							/>
						</div>
						<div style={css.item}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: "Image ID",
										value: deployment.image_id || "-",
									},
								]}
							/>
						</div>
						<div style={css.item}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: "Image Name",
										value: deployment.image_name || "-",
									},
								]}
							/>
						</div>
						<div style={css.item}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: "Tags",
										value: deployment.tags.length ? deployment.tags : "-",
									},
								]}
							/>
						</div>
					</div>
					<div className="layout horizontal flex">
						{editor}
					</div>
					<PodDeploymentEdit
						disabled={this.props.disabled}
						deployment={this.props.deployment}
						open={this.state.editOpen}
						onClose={this.onEditClose}
					/>
				</div>
			</Blueprint.Card>
		} else {
			return <Blueprint.Card
				key={deployment.id}
				compact={true}
				style={cardStyle}
			>
				<div className="layout vertical flex">
					<div className="layout horizontal flex" style={css.box}>
						<div className="layout center" style={css.checkBox}>
							<Blueprint.Checkbox
								style={css.check}
								checked={this.props.selected}
								onClick={(evt): void => {
									this.props.onSelect(evt.shiftKey)
								}}
							/>
						</div>
						<div style={css.itemFirst}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: "Deployment ID",
										value: deployment.id,
									},
									{
										label: "Commit ID",
										value: deployment.spec.substring(12),
										hover: specHover,
										valueClass: commitClass,
									},
								]}
							/>
							<button
								className="bp5-button bp5-small"
								style={css.cardButton}
								hidden={this.state.logsOpen}
								onClick={this.onLogsToggle}
							>Logs</button>
							<button
								className="bp5-button bp5-small bp5-active bp5-intent-danger"
								style={css.cardButton}
								hidden={!this.state.logsOpen}
								onClick={this.onLogsToggle}
							>Logs</button>
							<button
								className="bp5-button bp5-small"
								style={css.cardButton}
								hidden={this.state.editOpen}
								onClick={this.onEditToggle}
							>Settings</button>
							<button
								className="bp5-button bp5-small bp5-active bp5-intent-danger"
								style={css.cardButton}
								hidden={!this.state.editOpen}
								onClick={this.onEditToggle}
							>Settings</button>
						</div>
						<div style={css.itemMedium}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: "Zone",
										value: deployment.zone_name || "-",
									},
									{
										label: "Node",
										value: deployment.node_name || "-",
									},
									{
										label: "Instance",
										value: deployment.instance_name || "-",
									},
								]}
							/>
						</div>
						<div style={css.itemSmall}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: "State",
										value: stateValue,
										valueClass: stateClass,
									},
									{
										label: "Status",
										value: deployment.instance_status || "-",
										valueClass: statusClass,
									},
									domainInfo,
								]}
							/>
						</div>
						<div style={css.item}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: "Agent Status",
										value: MiscUtils.capitalize(
											deployment.instance_guest_status) || "-",
										valueClass: heartbeatClass,
									},
									{
										label: "Last Heartbeat",
										value: MiscUtils.formatSinceLocal(
											deployment.instance_heartbeat) || "-",
										hover: heartbeatHover,
										valueClass: heartbeatClass,
									},
									{
										label: "Uptime",
										value: deployment.instance_uptime || "-",
									},
								]}
							/>
						</div>
						<div style={css.item}>
							<PageInfo
								compact={true}
								style={css.info}
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
						<div style={css.itemFull}>
							<PageInfo
								compact={true}
								style={css.info}
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
						<div style={css.itemLast}>
							<PageInfo
								compact={true}
								style={css.info}
								bars={resourceBars}
							/>
							<button
								className="bp5-button bp5-small"
								style={css.cardButtonRight}
								onClick={(): void => {
									InstanceActions.filter({
										id: deployment.instance
									})
									Router.setLocation("/instances")
								}}
							>View Instance</button>
						</div>
					</div>
					<div className="layout horizontal flex">
						{editor}
					</div>
					<PodDeploymentEdit
						disabled={this.props.disabled}
						deployment={this.props.deployment}
						open={this.state.editOpen}
						onClose={this.onEditClose}
					/>
				</div>
			</Blueprint.Card>
		}
	}
}
