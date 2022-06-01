

export const resetTableData = (data: string) => {
  if (!data) return []
  const list = data && data.split('\n').filter((key: any)=> key)

  let dataSource: any = []
  if (Array.isArray(list)) {
    dataSource = list.map((item: any, i: any) => {
      const row = item.replace('[', '').replace(']', '').split(' ')
      // console.log('row:', row)

      return {
        id: i+ 1,
        status: row[0],
        name: row[1],
      }
    });
  }
  // console.log('dataSource:', dataSource)
  return dataSource
}