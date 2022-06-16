import React from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { Card, Alert, Typography } from 'antd';
import { useIntl, FormattedMessage, history } from 'umi';
import background_logo from '@/assets/logo_background.jpg';
import KeenTune_logo from '@/assets/KeenTune-logo.png';
import SelectLang from '@/components/RightContent/SelectLang';
import { useClientSize } from '@/uitls/uitls'
import Footer from '@/components/Footer';
import General from './general';
import styles from './index.less';

export default (): React.ReactNode => {
  const intl = useIntl();
  const { height } = useClientSize()
  const style = { background: `url(${background_logo}) no-repeat`, backgroundSize: '100% auto', minHeight: height }

  return (
    <div className={styles.home_root} style={style}>
      <div className={styles.content}>

        <div className={styles.title}>
          <img src={KeenTune_logo} className={styles.KeenTune_logo}/>
          <p>KeenTune（轻豚）是一款Linux上跨平台的AI算法与专家知识库双轮驱动的全栈调优工具，能够为系统研发、运维人员在单机、集群、分布式等不</p>
          <p style={{ marginBottom:0 }}>同系统环境提供轻量的敏感参数识别、智能参数调优、一键式专家调优的能力，也能够为算法人员提供可视化的算法优化服务。</p>
        </div>

        <General />

        <div className={styles.home_link}>
          <span>详细信息请访问：</span>
          <a href="http://keentune.io/home" target="_blank">http://keentune.io/home</a>
        </div>
      </div>

      <div className={styles.position_img}>
        <SelectLang className={styles.action} />
      </div>
    </div>
  );
};
