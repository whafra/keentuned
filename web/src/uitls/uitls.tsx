import React, { useEffect, useState, useLayoutEffect, useCallback } from 'react'
import { notification, message } from 'antd';
import { request } from 'umi';

export const myIsNaN = (value: any)=>{
    return (typeof value === 'number' && !isNaN(value)) || /^\d+$/.test(value);
  }

// 接口报信息处理
export const handleRes = (res: any, title?: string)=>{
  const reactNode = (
    <>
      {title}
      <div>
        {res?.msg?.split('\n').map((item: any, i: number)=>
          <div key={i} style={{textAlign:'left'}}>
            <pre style={{marginBottom:'0px' }}>{item}</pre>
          </div>
        )}
      </div>
    </>
  )
  
  if (res.suc) {
    message.success(reactNode);
  } else {
    message.error(reactNode);
  }
}

export const useClientSize = () => {
  const [ layout , setLayout ] = useState({ height : innerHeight , width : innerWidth })
  const onResize = useCallback(
      () => {
          setLayout({
              height: innerHeight,
              width: innerWidth
          })
      }, []
  )

  useEffect(() => {
      window.addEventListener('resize', onResize)
      return () => {
          window.removeEventListener('resize', onResize)
      }
  }, [])

  return layout
}

/**
 * 转换csv文件
 * @param data csv文件里的字符串
 * @returns { Array }
 */
export const dataDealWith = (data: string) => {
    const list = data && data.split('\n')
    // console.log('list:', list)
    const titleSet = list[0].split(',').map((item: any)=> item.toLowerCase())
    // console.log('titleSet:', titleSet)

    let dataSource = []
    if (Array.isArray(list)) {
        dataSource = list?.slice(1)?.filter((key: any)=> key).map((item: any, index: any)=> {
        let row: any = {}
        item.split(',').forEach((value: any, i: any) => {
            row.id = index + 1
            row[titleSet[i]] = value
        });
        return row
        })
    }
    return dataSource.reverse()
}

/**
 * 转换csv文件
 * @param data csv文件里的字符串
 * @returns { Array }
 */
 export const csvToJson = (data: string) => {
  const list = data && data.split('\n')

  let dataSource = []
  if (Array.isArray(list)) {
    const titleSet = list[0].split(',')
    dataSource = list?.slice(1)?.filter((key: any)=> key).map((item: any, index: any)=> {
      let row: any = {}
      item.split(',').forEach((value: any, i: any) => {
          row.id = index + 1
          row[titleSet[i]] = value
      });
      return row
    })
  }
  return dataSource
}

/**
 * 
 * @param data 
 * @returns 
 *     {
        "date": "2015-01-04",
        name: 
        value: 1,
        },
 */
export const timeFile = (data: string) => {
    const list = data && data.split('\n').filter(key=> key)
    // console.log('list:', list)

    let dataSource: any = []
    if (Array.isArray(list)) {
      // 去除文件列名行
      const dataList = list.slice(1)
      dataList?.forEach((item: any, i: any)=> {
          let algorithmItem: any = {}
          let benchmarkItem: any = {}
          let row = item.split(',').map((key: any)=> Number(key));
          algorithmItem = {
              date: i + 1,
              name: 'Algorithm time',
              value: row[1] - row[0] + row[3] - row[2]
          }
          benchmarkItem = {
              date: i + 1,
              name: 'Benchmark time',
              value: row[2] - row[1]
          }
          dataSource.push(algorithmItem);
          dataSource.push(benchmarkItem);
      })
    }
    // console.log('dataSource:', dataSource)s
    return dataSource;
}

/**
 * running（运行中），finish（完成），abort（用户终止），error
 */
export const statusEnum = {
    abort: { text: 'abort', status: 'Default' },
    running: { text: 'running', status: 'Processing' },
    finish: { text: 'finish', status: 'Success' },
    error: { text: 'error', status: 'Error' },
    //
    available: { text: 'available', status: 'Processing' },
    active: { text: 'active', status: 'Success' },
}

// 智能调优功能的算法
export const tuningAlgorithm = ['TPE', 'HORD', 'Random'] // 'Bayesian Optimize'
// 敏感参数功能算法
export const sensitivityAlgorithm = ['lasso', 'univariate', 'shap']

/**
 * @function http响应报错处理，报错弹框提示！
 **/ 
export const handleError = (response: any)=> {
  // 500报错优化
  if (response && response.status === 500) {
    notification.error({
      message: '校验错误, 请联系开发。',
    });
  } else if (response?.status && response.status !== 200) {
    const { status, url } = response;
    let arr = url.split('/');
    // 404报错， 文件 Not Found 提示！
    const meg = response.statusText? `${arr[arr.length -1].split('?')[0]} ${response.statusText}`: '查询数据失败！'
    message.error(meg, 3);

  } else if (!response) {
    notification.error({
      description: '您的网络发生异常，无法连接服务器',
      message: '网络异常',
    });
  }
}

/** 防抖 */
export const debounce = (fn: any)=> {
  let timeout: any = null;
  return (...args: any)=> { //...的作用：代表剩余参数；将数组展开作为函数的参数
    clearTimeout(timeout);
    timeout = setTimeout(() => {
      fn(...args) // fn.apply(this, arguments); 代表剩余参数
    }, 500);
  };
}

export const viewDetails = async (url: string, title: string) => {
  const data = await request(url, { skipErrorHandler: true, })
  const win: any = window.open('', '_blank')
  win.document.write(`<title>${title}</title>`)
  win.document.write(`<pre>${data}</pre>`)
  win.document.close()
}

/**
 * 
 * @param data 要保存的数据
 * @param name 文件后缀名
 */
export const saveToFile = (data: string, name: string) => {
  const a = document.createElement('a')

  // 定义文件名及后缀名
  let fileName = 'data.'+ name;
  a.download = fileName;
  a.style.display = 'none'

  // 生成一个blob的二进制数据，内容为json数据
  let blob = new Blob([data]); // JSON.stringify(data)
  console.log('blob:', blob)


  // 生成一个指向blob的url地址
  const url = URL.createObjectURL(blob);
  a.href = url
  console.log('url:', url)


  // body里生成一个a标签
  document.body.appendChild(a);
  a.click();

  // 移除a标签
  document.body.removeChild(a);

}
