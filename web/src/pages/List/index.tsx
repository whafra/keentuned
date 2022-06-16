import React from 'react';
import { PageLoading } from '@ant-design/pro-layout';
import { Card, Alert, Typography } from 'antd';
import { useIntl, FormattedMessage } from 'umi';
import styles from './index.less';

export default (props: any): React.ReactNode => { 
  const intl = useIntl();
  return (
    <div className={styles.list_root}>
      <div className={styles.content}>
         {props.children}
      </div>
    </div>
  );
};
