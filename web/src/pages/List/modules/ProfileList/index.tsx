import React, { useState, useEffect, useRef } from 'react';
import { Button, Tooltip, message, Space, Popconfirm } from 'antd';
import { FormattedMessage, useIntl, history } from 'umi';
import { DeleteOutlined, CopyOutlined, FormOutlined, UndoOutlined, HighlightOutlined } from '@ant-design/icons';
import type { ProColumns } from '@ant-design/pro-table';
import ProTable, { TableDropdown } from '@ant-design/pro-table';
//
import PageContainer from '@/components/public/PageContainer';
import PopoverEllipsis from '@/components/public/PopoverEllipsis';
import LogModal from '@/pages/List/LogModal'
import SetModal from './Set'
import CreateModal from './Create'
import Count from './Count'
import { handleRes, useClientSize, dataDealWith, statusEnum, viewDetails } from '@/uitls/uitls'
import { getRequestData } from './service'
import { resetTableData } from './dataDealWith'
import styles from './index.less';

// 不允许编辑的项
const defaultName = ['cpu_high_load.conf', 'io_high_throughput.conf', 'mysql_tpcc.conf',
 'net_high_throuput.conf', 'net_low_latency.conf']

export type TableListItem = {
  key: number;
  name: string;
  containers: number;
  creator: string;
  status: string;
  createdAt: number;
  memo: string;
};

/**
 * 静态调优列表
 */
export default () => {
  const { formatMessage } = useIntl();
  const [loading, setLoading] = useState(false)
  const [dataSource, setDataSource] = useState<any>([])
  const [listPage, setListPage] = useState({ 
    pageNum: 1,
    pageSize: 20,
    rows: [],
    total: 0,
  })
  const setModalRef: any = useRef(null)
  const createModalRef: any = useRef(null)
  const logModalRef: any = useRef(null)

  // 初始化状态
  const initialStatus = ()=> {
    setListPage({
      pageNum: 1,
      pageSize: 20,
      rows: [],
      total: 0,
    })
  }

  // 初始化请求数据
  const requestAllData = async (q?: any)=> {
    setLoading(true);
    try {
      const data = await getRequestData({ cmd: 'keentune profile list'}) || {}
      const resetData = resetTableData(data.msg)
      if (resetData && resetData.length) {
        setDataSource(resetData)
        // 前端分页
        setListPage({
          pageNum: 1,
          pageSize: listPage.pageSize,
          rows: resetData.slice(0, listPage.pageSize),
          total: resetData.length,
        });
      }
      setLoading(false);
    } catch (err) {
      setLoading(false);
    }
  }

  useEffect(()=> {
    requestAllData()
  }, [])

  // 前端分页
  const getTableData = (q: any)=> {
    const { total } = listPage
    const { pageNum, pageSize } = q
    if (dataSource && dataSource.length) {
      const start = (pageNum - 1) * pageSize
      setListPage({
        pageNum,
        pageSize,
        rows: dataSource.slice(start, start + pageSize),
        total,
      });
    } else {
      initialStatus()
    }
  }

  // 操作功能
  const cmdOperation = async (q: any, operateType: string)=> {
    operateType = formatMessage({ id: operateType })
    setLoading(true);
    try {
      const res = await getRequestData(q) || {}
      setLoading(false);
      if (res.suc) {
        requestAllData()
      } 
      handleRes(res, operateType)
    } catch (err) { 
      setLoading(false);
    }
  }

  const fn = (key: string, row: any)=> {
    switch (key) {
      case 'set': setModalRef.current?.show({ title: '设置调优项', url: '', row }); break
      case 'copy': createModalRef.current?.show({ title: key, row }); break
      case 'edit': createModalRef.current?.show({ title: key, row }); break
      case 'rollback': cmdOperation({ cmd: 'keentune profile rollback'}, key); break
      case 'delete': cmdOperation({ cmd: `keentune profile delete --name ${row.name}`}, key); break
      default: null
    }
  }

  const columns: ProColumns<TableListItem>[] = [
    {
      title: 'Profile Name',
      dataIndex: 'name',
      // ellipsis: true,
      width: 300,
      render: (text: any, row: any) => {
        return <PopoverEllipsis title={text} onClick={()=> logModalRef.current?.show({ title: 'Details', url: `/var/keentune/profile/${row.name}` })} />
      },
    },
    {
      title: 'Status',
      dataIndex: 'status',
      valueEnum: statusEnum,
    },
    {
      title: 'Target Group',
      dataIndex: 'target',
      render: (text, record) => <span><Count record={record} minWidth={50}/></span>,
    },
    {
      title: <FormattedMessage id="operations" />,
      key: 'option',
      width: 230,
      valueType: 'option',
      render: (text, record, _, action) => {
        const iconStyle = { fontSize:16,color:'#008dff' }
        const disableStyle = { fontSize:16,color:'#ccc' }
        return (
          <Space size={24}>
          {!defaultName.includes(record.name)?
            <Popconfirm
              title={ formatMessage({ id: 'confirm.content' }) }
              onConfirm={()=> fn('delete', record)}
              onCancel={()=>{}}
              okText={formatMessage({ id: 'btn.yes' }) }
              cancelText={formatMessage({ id: 'btn.no' }) }
            >
              <Tooltip placement="top" title={ formatMessage({ id: 'delete' }) }>
                <DeleteOutlined style={{ fontSize:16, }}/>
              </Tooltip>
            </Popconfirm>
            : <DeleteOutlined style={disableStyle}/> }

            <Tooltip placement="top" title={ formatMessage({ id: 'copy' }) }>
              <CopyOutlined onClick={()=> fn('copy', record)} style={iconStyle}/>
            </Tooltip>

            {!defaultName.includes(record.name)?
              <Tooltip placement="top" title={ formatMessage({ id: 'edit' }) }>
                <FormOutlined onClick={()=> fn('edit', record)} style={iconStyle}/>
              </Tooltip>
              : <FormOutlined style={disableStyle}/> }
            
            {record.status === 'active'?
              <Tooltip placement="top" title={ formatMessage({ id: 'rollback', defaultMessage: 'Rollback' }) }>
                <UndoOutlined onClick={()=> fn('rollback', record)} style={iconStyle}/>
              </Tooltip>
            : <UndoOutlined style={disableStyle}/> }
            
            <Tooltip placement="top" title={ formatMessage({ id: 'set' }) }>
              <HighlightOutlined onClick={()=> fn('set', record)} style={iconStyle}/>
            </Tooltip>
          </Space>
        )
      }
    },
  ];

  return (
    <div className={styles.static_page}>
      <PageContainer style={{ marginTop: 24,padding: 0 }}>
        <ProTable<TableListItem>
          loading={loading}
          headerTitle={<FormattedMessage id="pages.profileList.title" />}
          options={{ 
            reload: requestAllData,
            setting: true,
            density: false,
          }}
          size="small"
          columns={columns}
          dataSource={ listPage.rows }
          rowKey="id"
          pagination={{
            current: listPage.pageNum,
            pageSize: listPage.pageSize,
            total: listPage.total,
            size: "default",
            showSizeChanger: true,
            showTotal: (total, range) => { return `${formatMessage({id: 'total'})} ${total} ${formatMessage({id: 'records'})} ${listPage.pageNum} / ${Math.ceil(total / listPage.pageSize)} ${formatMessage({id: 'page'})}`},
            onChange: (page, pageSize) => { 
              const tempPage = pageSize !== listPage.pageSize? 1 : page
              getTableData({ pageNum: tempPage, pageSize })
            },
          }}
          search={false}
          dateFormatter="string"
          toolBarRender={() => [
            <Button key="button" type="primary" onClick={()=> createModalRef.current?.show({ title: 'create' })}>
              <FormattedMessage id="create" />
            </Button>,
          ]}
        />
        <SetModal ref={setModalRef} callback={requestAllData} />

        {/* Create */}
        <CreateModal ref={createModalRef} dataSource={dataSource} callback={requestAllData}/>
        {/* details */}
        <LogModal ref={logModalRef} />
      </PageContainer>
    </div>
  );
};