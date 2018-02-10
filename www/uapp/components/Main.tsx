/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ReactRouter from 'react-router-dom';
import * as Theme from '../Theme';
import Loading from './Loading';
import Nodes from './Nodes';
import * as NodeActions from '../actions/NodeActions';

interface State {
	disabled: boolean;
}

const css = {
	nav: {
		overflowX: 'auto',
		overflowY: 'hidden',
		userSelect: 'none',
	} as React.CSSProperties,
	link: {
		padding: '0 8px',
		color: 'inherit',
	} as React.CSSProperties,
	sub: {
		color: 'inherit',
	} as React.CSSProperties,
	heading: {
		marginRight: '11px',
		fontSize: '18px',
		fontWeight: 'bold',
	} as React.CSSProperties,
	loading: {
		margin: '0 5px 0 1px',
	} as React.CSSProperties,
};

export default class Main extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
		};
	}

	render(): JSX.Element {
		return <ReactRouter.HashRouter>
			<div>
				<nav className="pt-navbar layout horizontal" style={css.nav}>
					<div className="pt-navbar-group pt-align-left flex">
						<div className="pt-navbar-heading"
							style={css.heading}
						>Pritunl Zero</div>
						<Loading style={css.loading} size="small"/>
					</div>
					<div className="pt-navbar-group pt-align-right">
						<ReactRouter.Link
							className="pt-button pt-minimal pt-icon-layers"
							style={css.link}
							to="/nodes"
						>
							Nodes
						</ReactRouter.Link>
						<ReactRouter.Route render={(props) => (
							<button
								className="pt-button pt-minimal pt-icon-refresh"
								disabled={this.state.disabled}
								onClick={() => {
									let pathname = props.location.pathname;

									this.setState({
										...this.state,
										disabled: true,
									});

									if (pathname === '/nodes') {
										NodeActions.sync().then((): void => {
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
								}}
							>Refresh</button>
						)}/>
						<button
							className="pt-button pt-minimal pt-icon-log-out"
							onClick={() => {
								window.location.href = '/logout';
							}}
						>Logout</button>
						<button
							className="pt-button pt-minimal pt-icon-moon"
							onClick={(): void => {
								Theme.toggle();
								Theme.save();
							}}
						/>
					</div>
				</nav>
				<ReactRouter.Route path="/nodes" render={() => (
					<Nodes/>
				)}/>
			</div>
		</ReactRouter.HashRouter>;
	}
}
