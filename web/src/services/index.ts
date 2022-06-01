import { request } from 'umi'

function doGet(method: string) {
  return ['GET', 'get'].includes(method)
}

// 
export const requestData = async (method= 'GET', url: string, params: any ) => {
  const q = doGet(method)? { params }: { data: params }
  return request(`${url}`, { skipErrorHandler: true, timeout: 5000,
    method,
    ...q,
  })
}