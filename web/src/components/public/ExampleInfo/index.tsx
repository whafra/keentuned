
import React, { useState, useRef, useEffect } from 'react';
import styles from './index.less'

export const ExampleInfo = ({ content = '', rows = 4 }: any) => {
  const ellipsis: any = useRef(null)
	const [show, setShow] = useState(false)

  useEffect(() => {
		isEllipsis()
	}, [content]);

  const isEllipsis = () => {
		const clientHeight = ellipsis.current.clientHeight
		const scrollHeight = ellipsis.current.scrollHeight
		setShow(clientHeight < scrollHeight)
	};
  
  return (
    <div ref={ellipsis} className={styles.example_info} 
      style={{ height: rows*22, overflowY: show? 'scroll': 'unset',}}>
      <pre>{content}</pre>
    </div>
  )
}