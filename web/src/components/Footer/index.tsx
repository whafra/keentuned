import React from 'react';
import { useIntl } from 'umi';
import { DefaultFooter } from '@ant-design/pro-layout';
import styles from './index.less';

export default () => {
  // const intl = useIntl();
  // const currentYear = new Date().getFullYear();

  return (
    <div className={styles.layout_footer}>
       <span>2021 - 2022 Copyright  OpenAnolis</span>
    </div>
  );
};
