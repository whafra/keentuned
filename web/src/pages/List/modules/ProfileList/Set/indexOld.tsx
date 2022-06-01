
import React, { useCallback, useState, useEffect, forwardRef, useImperativeHandle } from 'react';
import { FormattedMessage, history, request } from 'umi'; 
import { Modal, message, Button, Space, Spin, Input, Table, Radio } from 'antd'
import { ExclamationCircleOutlined } from '@ant-design/icons'
import { debounce, dataDealError } from '@/uitls/uitls'
import { getRequestData } from '../service'
import styles from './index.less'

/**
 * set对话框
 */
export default forwardRef((props: any, ref: any) => {
  const [loading, setLoading] = useState<any>(false)
  const [visible, setVisible] = useState(false);
  const [data, setData] = useState<any>([]);
  const [listPage, setListPage] = useState<any>({ current: 1, pageSize: 10 })
  const [title, setTitle] = useState('');
  //
  const [selectValue, setSelectValue]: any = useState('');

  const initialStatus = ()=> {
    setVisible(false)
    setData([])
    setSelectValue('')
  }

  // 1.请求数据
  const getTableData = async (url: string) => {
    setLoading(true);
      try {
        const res = await request(url, { skipErrorHandler: true, })
        if (typeof res === 'string') {
          setData(res)
        } else if (res instanceof Object) {
          // json对象 转 格式化字符串。
          const tempData = JSON.stringify(res, null, 4)
          setData(tempData)
        }
        setLoading(false);
      } catch (err: any) {
        // console.log('err:', err)
        setLoading(false);
      }
  }

  useImperativeHandle(
    ref,
    () => ({
        show: ({ title= '', url, row={}}: any) => {
          setVisible(true)
          setTitle(title);
          url? getTableData(url): setData([row])
        }
    })
  )

  const onSearch = (value: string) => {
    // console.log('value:', value);
    // debounce
  }
  
  const columns = [
    {
      title: 'name',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
      render: (text: any, record: any)=> {
        return <div>
          <Radio value={record.id} checked={selectValue === record.id}
           onChange={(e)=> setSelectValue(e.target.value) }>{text}</Radio>
        </div>
      }
    },
  ]

  const getSet = async (name: string)=> {
    const q = { cmd: `keentune profile set ${name}`}
    setLoading(true);
    try {
      let res = await getRequestData(q) || {}
      setLoading(false);
      if (res.suc) {
        message.success('配置成功');
        // 初始化状态 && 刷新数据
        initialStatus()
        props.callback({})
      } else {
        dataDealError(res, '配置失败！')
      }
    } catch (err) {
      setLoading(false);
    }
  }
  const handleOk = () => {
    if (selectValue) {
      const name = data.filter((item: any) => item.id === selectValue)[0].name
      getSet(name)
    } else {
      message.error('请选择一个配置，再提交');
    }
  };
  const handleCancel = () => initialStatus();

	return (
      <Modal
        title={<>{title}</>}
        visible={visible}
        maskClosable={true}
        width={960}
        confirmLoading={loading}
        onCancel={handleCancel}
        footer={
          <Space>
            <Button onClick={handleCancel} style={{ padding: '0 32px'}}>取消</Button>
            <Button onClick={handleOk} type="primary" style={{ padding: '0 32px',marginRight:12}}>配置</Button>
          </Space>
        }
        bodyStyle={{ padding:'20px 30px',background:'#f8f8f8'}}
      >
        <Spin spinning={loading}>
          请选择需要的设置调优项： <Input.Search placeholder="搜索Group" onSearch={onSearch} style={{ width:'300px' }} />
          <div className={styles.set_div}>
            <Table className={styles.set_table}
              size="small" 
              dataSource={data}
              columns={columns}
              showHeader={false}
              rowKey="id"
              pagination={undefined}
            />
          </div>
        </Spin>
      </Modal>
	);
});