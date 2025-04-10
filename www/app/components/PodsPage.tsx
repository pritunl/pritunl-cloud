/// <reference path="../References.d.ts"/>
import * as React from 'react';
import PodsStore from '../stores/PodsStore';
import * as PodActions from '../actions/PodActions';

interface Props {
	minimal?: boolean;
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
	buttonMinimal: {
		userSelect: 'none',
		margin: '0 3px 0 0',
		width: "20px",
		minWidth: "20px",
	} as React.CSSProperties,
	buttonLast: {
		userSelect: 'none',
		margin: '0 0 0 0',
	} as React.CSSProperties,
	link: {
		cursor: 'pointer',
		userSelect: 'none',
		margin: '7px 5px 0 0',
	} as React.CSSProperties,
	linkMinimal: {
		cursor: 'pointer',
		userSelect: 'none',
		margin: '7px 2px 0 0',
	} as React.CSSProperties,
	current: {
		opacity: 0.5,
	} as React.CSSProperties,
	ellipsis: {
		margin: '7px 2px 0 0',
	} as React.CSSProperties,
};

export default class PodsPage extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			page: PodsStore.page,
			pageCount: PodsStore.pageCount,
			pages: PodsStore.pages,
			count: PodsStore.count,
		};
	}

	componentDidMount(): void {
		PodsStore.addChangeListener(this.onChange);
	}

	componentWillUnmount(): void {
		PodsStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			page: PodsStore.page,
			pageCount: PodsStore.pageCount,
			pages: PodsStore.pages,
			count: PodsStore.count,
		});
	}

	createPageButton = (pageNum: number): JSX.Element => {
		const page = this.state.page;
		const linkStyle = this.props.minimal ? css.linkMinimal : css.link;

		return (
			<span
				key={pageNum}
				style={page === pageNum ? {
					...linkStyle,
					...css.current,
				} : linkStyle}
				onClick={(): void => {
					PodActions.traverse(pageNum);
					if (this.props.onPage) {
						this.props.onPage();
					}
				}}
			>
				{pageNum + 1}
			</span>
		);
	}

	render(): JSX.Element {
		let page = this.state.page;
		let pages = this.state.pages;

		if (pages <= 1) {
			return <div/>;
		}

		let links: JSX.Element[] = [];

		if (this.props.minimal) {
			links.push(this.createPageButton(0));

			if (pages <= 3) {
				for (let i = 1; i < pages; i++) {
					links.push(this.createPageButton(i));
				}
			} else if (page <= 1) {
				if (page === 1) {
					links.push(this.createPageButton(1));
				}
				links.push(<span key="ellipsis1" style={css.ellipsis}>...</span>);
				links.push(this.createPageButton(pages - 1));
			} else if (page >= pages - 2) {
				links.push(<span key="ellipsis1" style={css.ellipsis}>...</span>);
				if (page === pages - 2) {
					links.push(this.createPageButton(pages - 2));
				}
				links.push(this.createPageButton(pages - 1));
			} else {
				links.push(<span key="ellipsis1" style={css.ellipsis}>...</span>);
				links.push(this.createPageButton(page));
				links.push(<span key="ellipsis2" style={css.ellipsis}>...</span>);
				links.push(this.createPageButton(pages - 1));
			}
		} else {
			let start = Math.max(0, page - 7);
			let end = Math.min(pages, start + 15);

			for (let i = start; i < end; i++) {
				links.push(this.createPageButton(i));
			}
		}

		return <div className="layout horizontal center-justified">
			<button
				className="bp5-button bp5-minimal bp5-icon-chevron-backward"
				hidden={pages < 5}
				disabled={page === 0}
				type="button"
				onClick={(): void => {
					PodActions.traverse(0);
					if (this.props.onPage) {
						this.props.onPage();
					}
				}}
			/>
			<button
				className="bp5-button bp5-minimal bp5-icon-chevron-left"
				style={this.props.minimal ? css.buttonMinimal : css.button}
				disabled={page === 0}
				type="button"
				onClick={(): void => {
					PodActions.traverse(Math.max(0, this.state.page - 1));
					if (this.props.onPage) {
						this.props.onPage();
					}
				}}
			/>
			{links}
			<button
				className="bp5-button bp5-minimal bp5-icon-chevron-right"
				style={this.props.minimal ? css.buttonMinimal : css.button}
				disabled={page === pages - 1}
				type="button"
				onClick={(): void => {
					PodActions.traverse(Math.min(
						this.state.pages - 1, this.state.page + 1));
					if (this.props.onPage) {
						this.props.onPage();
					}
				}}
			/>
			<button
				className="bp5-button bp5-minimal bp5-icon-chevron-forward"
				hidden={pages < 5}
				disabled={page === pages - 1}
				type="button"
				onClick={(): void => {
					PodActions.traverse(this.state.pages - 1);
					if (this.props.onPage) {
						this.props.onPage();
					}
				}}
			/>
		</div>;
	}
}
