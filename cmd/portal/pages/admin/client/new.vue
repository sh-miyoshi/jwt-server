<template>
  <div class="card">
    <div class="card-header">
      <h3>New Client Info</h3>
    </div>

    <div class="card-body">
      <div class="form-group row">
        <label for="id" class="col-sm-2 col-form-label">
          ID
          <span class="required">*</span>
        </label>
        <div class="col-md-5">
          <input
            v-model="id"
            type="text"
            class="form-control"
            :class="{ 'is-invalid': idValidateError }"
            @blur="validateClientID()"
          />
          <div class="invalid-feedback">
            {{ idValidateError }}
          </div>
        </div>
      </div>

      <div class="form-group row">
        <label for="accessType" class="col-sm-2 col-form-label">
          Access Type
        </label>
        <div class="col-md-5">
          <select v-model="accessType" name="accessType" class="form-control">
            <option>confidential</option>
            <option>public</option>
          </select>
        </div>
      </div>

      <div class="card-footer">
        <div v-if="error" class="alert alert-danger">
          {{ error }}
        </div>

        <button class="btn btn-primary mr-2" @click="create">Create</button>
        <nuxt-link to="/admin/client">Cancel</nuxt-link>
      </div>
    </div>
  </div>
</template>

<script>
import { v4 as uuidv4 } from 'uuid'
import { ValidateClientID } from '~/plugins/validation'

export default {
  data() {
    return {
      id: '',
      accessType: 'confidential',
      idValidateError: '',
      error: ''
    }
  },
  methods: {
    async create() {
      if (this.idValidateError.length > 0) {
        this.error = 'Please fix validation error before create.'
        return
      }

      let secret = ''
      if (this.accessType !== 'public') {
        secret = uuidv4()
      }

      const data = {
        id: this.id,
        access_type: this.accessType,
        secret
      }
      const projectName = this.$store.state.current_project
      const res = await this.$api.ClientCreate(projectName, data)
      console.log('client create result: %o', res)
      if (!res.ok) {
        this.error = res.message
        return
      }

      await this.$bvModal.msgBoxOk('successfully created.')
      this.$router.push('/admin/client')
    },
    validateClientID() {
      const res = ValidateClientID(this.id)
      if (!res.ok) {
        this.idValidateError = res.message
      } else {
        this.idValidateError = ''
      }
    }
  }
}
</script>

<style scoped>
.required {
  color: #ee2222;
  font-size: 18px;
}
</style>
