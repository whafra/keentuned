import { request } from 'umi'

export const getRequestData = async (params: any) => {
  return request('/cmd', { skipErrorHandler: true,
      method: 'post',
      data: params,
      timeout: 5000,
  })
}