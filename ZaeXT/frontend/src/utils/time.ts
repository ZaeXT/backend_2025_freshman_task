import 'dayjs/locale/zh-cn'

import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import utc from 'dayjs/plugin/utc'

dayjs.extend(relativeTime)
dayjs.extend(utc)

dayjs.locale('zh-cn')

export const formatDateTime = (value: string | number | Date, format = 'YYYY-MM-DD HH:mm') =>
  dayjs(value).format(format)

export const formatRelativeTime = (value: string | number | Date) => dayjs(value).fromNow()

export const toIsoString = (value: Date) => value.toISOString()
