// import { GlobalOutlined } from '@ant-design/icons';
import { Menu } from 'antd';
import { getLocale, setLocale, FormattedMessage } from 'umi';
import React from 'react';
import Cn_logo from '@/assets/Cn.png';
import En_logo from '@/assets/En.png';
import styles from './index.less';

const SelectLang = props => {

  const onClick = () => {
    const en = getLocale() == 'zh-CN' ? 'en-US': 'zh-CN'
    setLocale(en, false) // false不刷新页面
  }
  
  return (
    <div className={styles.lang} onClick={onClick}>
      <img src={getLocale() === 'zh-CN' ? En_logo : Cn_logo} />
    </div>
  )
};

export default SelectLang;
