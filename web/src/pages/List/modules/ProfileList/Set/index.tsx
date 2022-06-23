
import React, { useCallback, useState, useEffect, forwardRef, useImperativeHandle } from 'react';
import { FormattedMessage, history, request } from 'umi'; 
import { Modal, message, Button, Space, Spin, Input, Table, Radio } from 'antd'
import { ExclamationCircleOutlined } from '@ant-design/icons'
import { debounce, handleRes } from '@/uitls/uitls'
import { getRequestData } from '../service'
import styles from './index.less'

/**
 * set对话框
 */
export default forwardRef((props: any, ref: any) => {
  const [loading, setLoading] = useState<any>(false)
  const [visible, setVisible] = useState(false);
  const [rowData, setRowData] = useState<any>({});
  // 数据源
  const [data, setData] = useState<any>([]);
  const [title, setTitle] = useState('');
  //
  const [selectValue, setSelectValue]: any = useState('');

  const initialStatus = ()=> {
    setVisible(false)
    setRowData({})
    setSelectValue('')
  }

  // 1.请求数据
  const getTableData = async () => {
    setLoading(true);
      try {
        const res = await getRequestData({ cmd: "keentune config target" })
        if (res.suc) {
          // ...提取group
          const tempData = res?.msg?.match(/\[target-group-.*?\]/gi) || []
          setData(tempData.map((item: any, i:number)=> `group${i+1}`) )
        }
        setLoading(false);
      } catch (err: any) {
        setLoading(false);
      }
  }

  useImperativeHandle(
    ref,
    () => ({
        show: ({ title= '', row={}}: any) => {
          setVisible(true)
          setTitle(title);
          setRowData(row)
          getTableData()
        }
    })
  )

  // 提交
  const getSet = async (name: string)=> {
    const q = { cmd: `keentune profile set --${name} ${rowData.name}`}
    setLoading(true);
    try {
      let res = await getRequestData(q) || {}
      setLoading(false);
      if (res.suc) {
        handleRes(res, '配置成功')
        // 初始化状态 && 刷新数据
        initialStatus()
        props.callback({})
      } else {
        handleRes(res, '配置失败')
      }
    } catch (err) {
      setLoading(false);
    }
  }

  const handleOk = () => {
    if (selectValue) {
      getSet(selectValue)
    } else {
      message.error('请选择一个配置，再提交');
    }
  };
  const handleCancel = () => initialStatus();

	return (
      <Modal
        title={<Space><ExclamationCircleOutlined style={{ fontSize:14,color:'#008dff'}}/>{title}</Space>}
        visible={visible}
        maskClosable={true}
        width={960}
        confirmLoading={loading}
        onCancel={handleCancel}
        footer={
          <Space>
            <Button onClick={handleCancel} style={{ padding: '0 32px'}}>Cancel</Button>
            <Button onClick={handleOk} type="primary" style={{ padding: '0 32px',marginRight:12}}>Set</Button>
          </Space>
        }
        bodyStyle={{ padding:'20px 30px',background:'#f8f8f8'}}
      >
        <Spin spinning={loading}>
          <div className={styles.set_label}>请选择需要设置的group：</div>
          <div className={styles.set_div}>
            {data?.map((item: any, i: any)=> (
              <div key={i} className={styles.set_item}>
                <Radio value={item} checked={selectValue === item}
                  onChange={(e)=> setSelectValue(e.target.value) }>{item}</Radio>
              </div>
            ) )}
          </div>
        </Spin>
      </Modal>
	);
});