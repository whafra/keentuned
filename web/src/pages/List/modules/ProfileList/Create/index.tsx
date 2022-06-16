
import React, { useCallback, useState, useEffect, forwardRef, useImperativeHandle } from 'react';
import { FormattedMessage, useIntl, history, request } from 'umi'; 
import { Modal, message, Button, Space, Spin, Form, Input, InputNumber, Tooltip, Select, Row, Col } from 'antd'
import { ExclamationCircleOutlined } from '@ant-design/icons'
import CodeEditer from '@/components/public/CodeEditer';
import { ExampleInfo } from '@/components/public/ExampleInfo';
import { requestData } from '@/services';
import styles from './index.less'
import { isNull } from 'lodash';

/**
 * 对话框
 */
export default forwardRef((props: any, ref: any) => {
  const { formatMessage } = useIntl();
  const { dataSource = [] } = props
  const [loading, setLoading] = useState<any>(false)
  const [visible, setVisible] = useState(false);
  // 
  const [title, setTitle] = useState('');
  const [data, setData] = useState<any>('');
  const [form] = Form.useForm();
  // 展示示例
  const [showTargetExample, setShowTargetExample]= useState<any>(false) 

  // 1.请求数据
  const getRequestData = async (url: string) => {
    setLoading(true);
      try {
        const res = await request(url, { skipErrorHandler: true, })
        if (typeof res === 'string') {
          form.setFieldsValue({ info: res })
        } 
        // else if (res instanceof Object) {
        //   // json对象 转 格式化字符串。
        //   const jsonStr = JSON.stringify(res, null, 4)
        //   // setData(jsonStr)
        //   form.setFieldsValue({ info: jsonStr })
        // }
        setLoading(false);
      } catch (err: any) {
        setLoading(false);
      }
  }
  
  useImperativeHandle(
    ref,
    () => ({
        show: ({ title= '', row }: any) => {
          setVisible(true)
          setTitle(title);
          if (row) {
            const tempName = row?.name?.split('.conf')[0]
            form.setFieldsValue({
              name: title === 'copy'? `${tempName}_copy`: tempName,
              // info: row.target
            })
            getRequestData(`/var/keentune/profile/${row.name}`)
          }
        }
    })
  )

  // 表单提交
  const getFormData = () => {
    setLoading(true);
    form.validateFields().then(async ( values ) => {
      const { name } = values
      const { suc, msg } = await requestData('POST', '/write', { ...values, name:`${name}.conf` })
      if (suc) {
        // 重置状态 && 跳转页面
        initialStatus()
        props.callback()
      } else {
        message.error(msg || '请求错误')
      }
       setLoading(false);
    }).catch(( err ) => {
      setLoading(false);
    })
  }

  const initialStatus= ()=> {
    form.resetFields();
    setVisible(false);
  }

  const handleCancel = () => {
    initialStatus()
  };
  const handleOk = () => {
    if (!showTargetExample) {
      getFormData()
    } else {
      setShowTargetExample(0)
      // getFormData()
    }
  };
  useEffect(()=> {
    if (showTargetExample === 0) getFormData()
  }, [showTargetExample])

  const validatorFn = (_: any, value: any) => {
    if (!value) { return Promise.resolve() }
    const list = value.split('\n');
     
    // 校验每一行的格式
    let validate = true
    let row = 0
    for (let item of list) {
      ++row
      if (item.trim() ==='') {
        validate = true
      } else if (item.match(/^\[.*?\]$/g)) {
        validate = true
      } else if (item.trim().split(':')?.length === 2 && item.trim().split(':')[0] && item.trim().split(':')[1]) {
        validate = true
      } else {
        validate = false
        break
      }
    }
    return validate ? Promise.resolve(): Promise.reject(new Error(
      `${formatMessage({id: 'ProfileList.validateInfo1'})} ${row} ${formatMessage({id: 'ProfileList.validateInfo2'})}`
    ))
  }

  const label_Target = (
    <div className={styles.variableLabel}>
      <span>Content details</span>
      <span className={styles.Bulk_btn} onClick={()=> setShowTargetExample(!showTargetExample)}>
        {showTargetExample ? <FormattedMessage id="fill.in" />: <FormattedMessage id="examples" />}
      </span>
    </div>
  )
	return (
      <Modal
        title={<Space>
          <ExclamationCircleOutlined style={{ fontSize:14,color:'#008dff'}}/><FormattedMessage id={title} />
        </Space>
        }
        visible={visible}
        maskClosable={false}
        width={900}
        confirmLoading={loading}
        onCancel={handleCancel}
        footer={
          <Space>
            <Button onClick={handleCancel}><FormattedMessage id="btn.close" /></Button> 
            <Button type="primary" onClick={handleOk} style={{marginRight:12}}>
              <FormattedMessage id="btn.ok" />
            </Button> 
          </Space>
        }
        bodyStyle={{ padding: '20px 30px',background: '#f8f8f8'}}
      >
        <Spin spinning={loading}>
          <Form form={form} layout="vertical">
              <Form.Item label="Profile Name"
                name="name"
                rules={[
                  {required: true, message: formatMessage({id: 'Input.placeholder'})},
                  {validator: (_, value) =>
                    (title !== 'edit' && value && dataSource?.filter((item: any)=> item.name === `${value}.conf`).length ) ?
                      Promise.reject(new Error('Profile Name名字重复!')): Promise.resolve(),
                  }
                ]}
              >
                <Input placeholder={formatMessage({id:'Input.placeholder'})} addonAfter=".conf" autoComplete="off" />
              </Form.Item>
             
              {showTargetExample?
              <Form.Item label={label_Target}
              className={styles.last_form_Item}
              name="target"
              rules={[
                { required: true },
              ]}>
              <ExampleInfo rows={8}
                content={`[my.cnf]
innodb_file_per_table: 1
lower_case_table_names: 1
sort_buffer_size : "10G"
max_allowed_packet : "16M"
tmp_table_size : "10G"
innodb_buffer_pool_size : "100G"
max_connections : 10000
key_buffer_size : "32M"
key_cache_block_size : "32M"
key_cache_age_threshold : 300
read_buffer_size : 1048576
read_rnd_buffer_size : "256K"
table_definition_cache : 2048
table_open_cache_instances : 16
open_files_limit : 65535
table_open_cache : 2048
general_log : "OFF"
log_bin : "OFF"
autocommit : "OFF"
innodb_lock_wait_timeout : 50
innodb_open_files : 3000
innodb_spin_wait_delay : 6
innodb_table_locks : "OFF"
skip_ssl : "ON"
innodb_buffer_pool_instances: 8
innodb_log_file_size : "32G"
innodb_log_file_instances : 4
performance_schema : "OFF"

[RamDisk]
/mnt/mysqldata: "204800m"

[benchmarksql]
warehouses: 100
loadWorkers: 200

[sysctl]
vm.swappiness: 1
vm.dirty_ratio : 5
kernel.sched_cfs_bandwidth_slice_us : 21000
kernel.sched_migration_cost_ns : 1381000
kernel.sched_latency_ns : 16110000
kernel.sched_min_granularity_ns : 8250000
kernel.sched_nr_migrate : 53
kernel.sched_wakeup_granularity_ns : 50410000
kernel.sched_autogroup_enabled : 0
kernel.numa_balancing : 0
net.core.rmem_default : 21299200
net.core.rmem_max : 21299200
net.core.wmem_default : 21299200
net.core.wmem_max : 21299200
net.ipv4.tcp_rmem : "40960 8738000 62914560"
net.ipv4.tcp_wmem : "40960 8738000 62914560"
net.core.dev_weight : 97
net.ipv4.tcp_max_syn_backlog : 20480
net.core.somaxconn : 1280
net.ipv4.tcp_max_tw_buckets : 360000
              `}
              />
            </Form.Item>
            :
            <Form.Item label={label_Target}
              className={styles.last_form_Item}
              name="info"
              rules={[
                {required: true, message: formatMessage({id: 'Input.placeholder'})},
                {validator: validatorFn }
              ]}
            >
              <Input.TextArea rows={8} placeholder={formatMessage({id: 'TextArea.placeholder'})} onChange={()=> {}} />
            </Form.Item>
          }
          </Form>
        </Spin>
      </Modal>
	);
});