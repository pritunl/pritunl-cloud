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
	box: {
		flex: 1,
		minWidth: "280px",
		margin: "10px 10px 20px 10px",
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
						title="No deployments"
						description="Update unit spec to create deployments."
						noDelay={true}
					/>
				</div>
			</div>
		}

		deployments.forEach((deployment: ServiceTypes.Deployment): void => {
			let publicIps = deployment.public_ips
			if (!publicIps || !publicIps.length) {
				publicIps = ["-"]
			}

			let publicIps6 = deployment.public_ips6
			if (!publicIps6 || !publicIps6.length) {
				publicIps6 = ["-"]
			}

			let privateIps = deployment.private_ips
			if (!privateIps || !privateIps.length) {
				privateIps = ["-"]
			}

			let privateIps6 = deployment.private_ips6
			if (!privateIps6 || !privateIps6.length) {
				privateIps6 = ["-"]
			}

			cards.push(<Blueprint.Card
				key={deployment.id}
				compact={true}
				style={css.card}
			>
				<div className="layout horizontal">
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
									label: "Status",
									value: deployment.instance_status || "-",
								},
								{
									label: "State",
									value: MiscUtils.capitalize(deployment.state) || "-",
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
				</div>
			</Blueprint.Card>)
		})

		return <div className="layout horizontal wrap">
			<div style={css.box}>
				<Blueprint.CardList>
					{cards}
				</Blueprint.CardList>
			</div>
		</div>
	}
}
