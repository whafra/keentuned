import React, { useState, useRef, useEffect } from "react";
import { Tooltip } from "antd"
import { request, history } from 'umi';
import styles from './index.less';

const PopoverEllipsis = ({ title, width, children, linkTo, onClick }: any) => {

	const ellipsis: any = useRef(null)
	const [show, setShow] = useState(false)

  useEffect(() => {
		isEllipsis()
	}, [title]);

	const isEllipsis = () => {
		const clientWidth = ellipsis.current.clientWidth
		const scrollWidth = ellipsis.current.scrollWidth
		setShow(clientWidth < scrollWidth)
	};

	const clickText = (onClick ? <span className={styles.ellipsis_click} onClick={onClick}>{title}</span> : title)
	// 判断文本内容是否有链接跳转
	const linkText = (linkTo ? <a target="_blank" href={linkTo}>{title}</a> : clickText)
	// 判断是否有文本内容
	const text = (title ? linkText : '-')

	const renderDom = (
		<div ref={ellipsis} className={styles.ellipsis} style={{ width: width || '100%' }}>
			{children ? children : title||'-'}
		</div>
	)

	return (
		show?
			<Tooltip placement="topLeft" title={title}>
				<div ref={ellipsis} className={styles.ellipsis} style={{ width: width || '100%' }}>
					{children ? children : text||'-'}
				</div>
			</Tooltip>
      :
			<div ref={ellipsis} className={styles.ellipsis} style={{ width: width || '100%' }}>
				{children ? children : text||'-'}
			</div>
	);
};

export default PopoverEllipsis;


