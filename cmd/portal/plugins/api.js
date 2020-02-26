import axios from 'axios'
import { AuthHandler } from './auth'

class APIClient {
  async _request(url, method, data) {
    const h = new AuthHandler()
    let res = await h.GetToken()

    if (!res.ok) {
      return res
    }

    const headers = {
      Authorization: 'Bearer ' + res.accessToken
    }

    if (!data) {
      headers['Content-Type'] = 'application/json'
    }

    const handler = axios.create({
      headers,
      timeout: 10000
    })

    try {
      switch (method) {
        case 'GET':
          res = await handler.get(url)
          break
        case 'POST':
          res = await handler.post(url, data)
          break
        case 'PUT':
          res = await handler.put(url, data)
          break
        case 'DELETE':
          res = await handler.delete(url)
          break
        default:
          return {
            ok: false,
            message: 'HTTP Method ' + method + ' is unsupported',
            statusCode: 500
          }
      }
      return { ok: true, data: res.data }
    } catch (error) {
      console.log(error)
      if (error.response) {
        if (error.response.status >= 400 && error.response.status < 500) {
          return {
            ok: false,
            message: 'auth required',
            statusCode: error.response.status
          }
        }
      }
      return {
        ok: false,
        message: 'Failed to request the server',
        statusCode: 500
      }
    }
  }

  async ProjectGetList() {
    const url = process.env.SERVER_ADDR + '/api/v1/project'
    const res = await this._request(url, 'GET')
    return res
  }

  async ProjectCreate(projectName) {
    // TODO(set all param)
    const url = process.env.SERVER_ADDR + '/api/v1/project'
    const res = await this._request(url, 'POST', { name: projectName })
    return res
  }
}

export default (context, inject) => {
  inject('api', new APIClient(context))
}