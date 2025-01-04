/// <reference path="../References.d.ts"/>
import * as React from "react"
import * as Blueprint from "@blueprintjs/core"
import * as Theme from '../Theme';
import Help from "./Help"
import * as PodTypes from "../types/PodTypes"
import * as PodActions from "../actions/PodActions"
import * as MiscUtils from '../utils/MiscUtils';
import NonState from './NonState';
import PageInfo from "./PageInfo"
import PodDeployment from "./PodDeployment"
import * as PageInfos from './PageInfo';

interface Props {
	hidden: boolean
	disabled?: boolean
	selected: Selected
	lastSelected: string
	onSelect: (selected: Selected, lastSelected: string) => void
	unit: PodTypes.PodUnit
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

export default class PodUnit extends React.Component<Props, State> {
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
		let deployments: PodTypes.Deployment[]
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

		deployments.forEach((deployment: PodTypes.Deployment): void => {
			cards.push(<PodDeployment
				key={deployment.id}
				hidden={this.props.hidden}
				disabled={this.props.disabled}
				selected={!!this.props.selected[deployment.id]}
				commitMap={commitMap}
				deployment={deployment}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.props.selected,
					};

					if (shift) {
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
			/>)
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
