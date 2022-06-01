import React, { useState } from 'react';
import { Card, Alert, Typography } from 'antd';
import { useIntl, FormattedMessage, history } from 'umi';
import list_2 from '@/assets/list_2.png';
import list_3 from '@/assets/list_3.png';
import list_4 from '@/assets/list_4.jpg';
import styles from './index.less';

export default () => {
  // 判断用户是否已下载插件
  const [tunDownloaded, setTunDownloaded] = useState(false); //true 
  const [senDownloaded, setSenDownloaded] = useState(false); //true
  
  const Template = ({ route='', logo='', title='', text='' })=> (
    <div className={styles.item} onClick={()=> { history.push(route) }}>
      <div className={styles.logo}>
        <img src={logo} />
      </div>
      <p className={styles.title}>{title}</p>
      <div className={styles.text}>{text}</div>
    </div>
  )

  const Template2 = ({ route='', logo='', title='', text='' })=> (
    <div className={styles.item_disable}>
      <div className={styles.logo}>
        <img src={logo} />
      </div>
      <p className={styles.title}>{title}</p>
      <div className={styles.text}>{text}</div>
      {/** 下载提示 */}
      <div className={styles.mask_layout}>
        <div className={styles.mask_context}>
          <div className={styles.text}>插件下载说明</div>
          <div className={styles.button}>操作文档链接</div>
        </div>
      </div>
    </div>
  )

  
  return (
    <div className={styles.general}>

      <div className={styles.list}>
         <div className={styles.item} onClick={()=> { history.push('/list/static-page')}}>
          <div className={styles.logo}>
            <img src={list_4} />
          </div>
          <p className={styles.title}>一键式专家调优</p>
          <div className={styles.text}>
           积累多种场景下业务全栈调优的专家知识，根据业务类型对系统进行一键式优化
          </div>
        </div>

        {tunDownloaded?
          <Template
            route="/list/tuning-task"
            logo={list_2}
            title='智能参数调优'
            text='提供多种高校算法，根据业务运行状态智能化调整系统全栈参数，使业务运行在最佳环境中' 
          /> 
          : 
          <Template2
            logo={list_2}
            title='智能参数调优'
            text='提供多种高校算法，根据业务运行状态智能化调整系统全栈参数，使业务运行在最佳环境中' 
          />
        }

        {senDownloaded?
          <Template
            route="/list/sensitive-parameter"
            logo={list_3}
            title='敏感参数识别'
            text='对参数进行建模识别出对结果影响度高的参数，并提供优选值范围，辅助参数可解释性和优化' 
          />
        :
          <Template2
            logo={list_3}
            title='敏感参数识别'
            text='对参数进行建模识别出对结果影响度高的参数，并提供优选值范围，辅助参数可解释性和优化' 
          />
        }

      </div>
    </div>

  );
};
