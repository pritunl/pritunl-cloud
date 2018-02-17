/// <reference path="../References.d.ts"/>
import * as React from 'react';
import InstancesStore from '../stores/InstancesStore';
import * as InstanceActions from '../actions/InstanceActions';

interface Props {
	onPage?: () => void;
}

interface State {
	page: number;
	pageCount: number;
	pages: number;
	count: number;
}

const css = {
	button: {
		userSelect: 'none',
		margin: '0 5px 0 0',
	} as React.CSSProperties,
	buttonLast: {
		userSelect: 'none',
		margin: '0 0 0 0',
	} as React.CSSProperties,
	link: {
		cursor: 'pointer',
		userSelect: 'none',
		margin: '5px 5px 0 0',
	} as React.CSSProperties,
	current: {
		opacity: 0.5,
	} as React.CSSProperties,
};

export default class InstancesPage extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			page: InstancesStore.page,
			pageCount: InstancesStore.pageCount,
			pages: InstancesStore.pages,
			count: InstancesStore.count,
		};
	}

	componentDidMount(): void {
		InstancesStore.addChangeListener(this.onChange);
	}

	componentWillUnmount(): void {
		InstancesStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			page: InstancesStore.page,
			pageCount: InstancesStore.pageCount,
			pages: InstancesStore.pages,
			count: InstancesStore.count,
		});
	}

	render(): JSX.Element {
		let page = this.state.page;
		let pages = this.state.pages;

		if (pages <= 1) {
			return <div/>;
		}

		let offset = 1;
		if (pages < 5) {
			offset = 0;
		}

		let links: JSX.Element[] = [];
		let start = Math.max(offset, page - 7);
		let end = Math.min(pages - offset, start + 15);

		for (let i = start; i < end; i++) {
			links.push(<span
				key={i}
				style={page === i ? {
					...css.link,
					...css.current,
				} : css.link}
				onClick={(): void => {
					InstanceActions.traverse(i);
					if (this.props.onPage) {
						this.props.onPage();
					}
				}}
			>
				{i + 1}
			</span>);
		}

		return <div className="layout horizontal center-justified">
			<button
				className="pt-button pt-minimal pt-icon-chevron-backward"
				hidden={!offset}
				disabled={page === 0}
				type="button"
				onClick={(): void => {
					InstanceActions.traverse(0);
					if (this.props.onPage) {
						this.props.onPage();
					}
				}}
			/>
			<button
				className="pt-button pt-minimal pt-icon-chevron-left"
				style={css.button}
				disabled={page === 0}
				type="button"
				onClick={(): void => {
					InstanceActions.traverse(Math.max(0, this.state.page - 1));
					if (this.props.onPage) {
						this.props.onPage();
					}
				}}
			/>
			{links}
			<button
				className="pt-button pt-minimal pt-icon-chevron-right"
				style={css.button}
				disabled={page === pages - 1}
				type="button"
				onClick={(): void => {
					InstanceActions.traverse(Math.min(
						this.state.pages - 1, this.state.page + 1));
					if (this.props.onPage) {
						this.props.onPage();
					}
				}}
			/>
			<button
				className="pt-button pt-minimal pt-icon-chevron-forward"
				hidden={!offset}
				disabled={page === pages - 1}
				type="button"
				onClick={(): void => {
					InstanceActions.traverse(this.state.pages - 1);
					if (this.props.onPage) {
						this.props.onPage();
					}
				}}
			/>
		</div>;
	}
}
