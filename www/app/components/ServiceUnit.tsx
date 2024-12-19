/// <reference path="../References.d.ts"/>
import * as React from "react"
import * as Blueprint from "@blueprintjs/core"
import * as Theme from '../Theme';
import Help from "./Help"
import * as ServiceTypes from "../types/ServiceTypes"
import * as ServiceActions from "../actions/ServiceActions"
import * as MiscUtils from '../utils/MiscUtils';
import NonState from './NonState';
import PageInfo from "./PageInfo"
import * as PageInfos from './PageInfo';

interface Props {
	hidden: boolean
	disabled?: boolean
	selected: Selected
	lastSelected: string
	onSelect: (selected: Selected, lastSelected: string) => void
	unit: ServiceTypes.ServiceUnit
}

interface State {
}

interface Selected {
	[key: string]: boolean
}

const css = {
	container: {
		height: "900px",
		overflowY: "auto",
		marginBottom: "10px",
	} as React.CSSProperties,
	box: {
		flex: 1,
		minWidth: "280px",
		margin: "10px",
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
	} as React.CSSProperties,
	itemFirst: {
		flex: "1 1 auto",
		minWidth: "100px",
		margin: "0 5px 0 -4px",
	} as React.CSSProperties,
	item: {
		flex: "1 1 auto",
		minWidth: "30px",
		margin: " 0 5px",
	} as React.CSSProperties,
	itemLast: {
		flex: "0 1 auto",
		minWidth: "30px",
		margin: " 0 5px",
	} as React.CSSProperties,
	specHover: {
		padding: "10px",
		fontSize: "12px",
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
}

export default class ServiceUnit extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context)
		this.state = {
		}
	}

	render(): JSX.Element {
		if (this.props.hidden) {
			return <div></div>
		}

		let label = "deployment"
		let labelTitle = "Deployment"
		if (this.props.unit.kind == "image") {
			label = "image"
			labelTitle = "Image"
		}

		let cards: JSX.Element[] = []
		let deployments: ServiceTypes.Deployment[]
		if (this.props.unit && this.props.unit.deployments) {
			deployments = this.props.unit.deployments
		} else {
			deployments = []
		}

		if (!deployments.length) {
			return <div className="layout horizontal wrap">
				<div style={css.boxEmpty}>
					<NonState
						hidden={false}
						iconClass="bp5-icon-dashboard"
						title={"No " + label + "s"}
						description={"Update unit spec to create " +
							label + "s."}
						noDelay={true}
					/>
				</div>
			</div>
		}

		let commitMap: Record<string, number> = {}
		if (this.props.unit.commits) {
			let count = 0
			for (let commit of this.props.unit.commits) {
				commitMap[commit.id] = count
				count -= 1
			}
		}

		deployments.forEach((deployment: ServiceTypes.Deployment): void => {
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
				case "archived":
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

			let commitIndex = commitMap[deployment.spec]
			let commitClass = commitIndex === 0 ? "bp5-text-intent-success" :
				"bp5-text-intent-danger"

				let specHover = <div
				className="bp5-content-popover"
				style={css.specHover}
			>
				<PageInfo
					compact={true}
					style={css.info}
					fields={[
						{
							label: "Commit",
							value: deployment.spec.substring(0, 24) || "-",
						},
						{
							label: "Tag",
							value: "1.2.5823.344",
						},
						{
							label: "Behind",
							value: commitIndex,
							valueClass: commitClass,
						},
					]}
				/>
			</div>

			let deplyStatus = MiscUtils.capitalize(deployment.status) || "-"
			let heartbeatClass = ""
			if (deployment.status === "healthy") {
				heartbeatClass = "bp5-text-intent-success"
			} else {
				heartbeatClass = "bp5-text-intent-danger"
			}

			let heartbeatHover = <div
				className="bp5-content-popover"
				style={css.specHover}
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

			if (deployment.kind === "image" && deployment.image_id) {
				cards.push(<Blueprint.Card
					key={deployment.id}
					compact={true}
					style={cardStyle}
				>
					<div className="layout horizontal flex">
						<div className="layout center" style={css.checkBox}>
							<Blueprint.Checkbox
								style={css.check}
								checked={!!this.props.selected[deployment.id]}
								onClick={(evt): void => {
									let selected = {
										...this.props.selected,
									};

									if (evt.shiftKey) {
										let deployments = this.props.unit.deployments;
										let start: number;
										let end: number;

										for (let i = 0; i < deployments.length; i++) {
											let deply = deployments[i];

											if (deply.id === deployment.id) {
												start = i;
											} else if (deply.id === this.props.lastSelected) {
												end = i;
											}
										}

										if (start !== undefined && end !== undefined) {
											if (start > end) {
												end = [start, start = end][0];
											}

											for (let i = start; i <= end; i++) {
												selected[deployments[i].id] = true;
											}

											this.props.onSelect(selected, deployment.id)

											return;
										}
									}

									if (selected[deployment.id]) {
										delete selected[deployment.id];
									} else {
										selected[deployment.id] = true;
									}

									this.props.onSelect(selected, deployment.id)
								}}
							/>
						</div>
						<div style={css.itemFirst}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: "Image ID",
										value: deployment.id,
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
										label: "Commit ID",
										value: deployment.spec.substring(0, 12),
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
										label: "Timestamp",
										value: MiscUtils.formatDateLocal(
											deployment.instance_heartbeat) || "-",
									},
								]}
							/>
						</div>
					</div>
				</Blueprint.Card>)
			} else {
				cards.push(<Blueprint.Card
					key={deployment.id}
					compact={true}
					style={cardStyle}
				>
					<div className="layout horizontal flex">
						<div className="layout center" style={css.checkBox}>
							<Blueprint.Checkbox
								style={css.check}
								checked={!!this.props.selected[deployment.id]}
								onClick={(evt): void => {
									let selected = {
										...this.props.selected,
									};

									if (evt.shiftKey) {
										let deployments = this.props.unit.deployments;
										let start: number;
										let end: number;

										for (let i = 0; i < deployments.length; i++) {
											let deply = deployments[i];

											if (deply.id === deployment.id) {
												start = i;
											} else if (deply.id === this.props.lastSelected) {
												end = i;
											}
										}

										if (start !== undefined && end !== undefined) {
											if (start > end) {
												end = [start, start = end][0];
											}

											for (let i = start; i <= end; i++) {
												selected[deployments[i].id] = true;
											}

											this.props.onSelect(selected, deployment.id)

											return;
										}
									}

									if (selected[deployment.id]) {
										delete selected[deployment.id];
									} else {
										selected[deployment.id] = true;
									}

									this.props.onSelect(selected, deployment.id)
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
										value: deployment.spec.substring(0, 12),
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
						<div style={css.item}>
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
										label: 'Public IPv4',
										value: publicIps,
									},
									{
										label: 'Private IPv4',
										value: privateIps,
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
										label: 'Public IPv6',
										value: publicIps6,
									},
									{
										label: 'Private IPv6',
										value: privateIps6,
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
						</div>
					</div>
				</Blueprint.Card>)
			}
		})

		return <div className="layout horizontal wrap" style={css.container}>
			<div style={css.box}>
				<Blueprint.CardList>
					{cards}
				</Blueprint.CardList>
			</div>
		</div>
	}
}
