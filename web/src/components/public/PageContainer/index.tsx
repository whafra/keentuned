import React from 'react';
import styles from './index.less';

const ContentContainer = (props: any) => {
  const { style={}, title, height } = props;
  return (
    <div className={styles.page_card_container} style={{ minHeight: height || 'unset', ...style}}>
      {title ? <div className={styles.page_title}>{title}</div>: null}
      {props.children}
    </div>
  );
}

export default ContentContainer
