<template>
  <div class="card">
    <div class="card-header">
      <h3>New Project Info</h3>
    </div>

    <div class="card-body">
      <div class="form-group row">
        <label for="name" class="col-sm-2 col-form-label">
          Name
          <span class="required">*</span>
        </label>
        <div class="col-md-5">
          <input
            v-model="name"
            type="text"
            class="form-control"
            :class="{ 'is-invalid': nameValidateError }"
          />
          <div class="invalid-feedback">
            {{ nameValidateError }}
          </div>
        </div>
      </div>

      <div class="card-footer">
        <div v-if="error" class="alert alert-danger">
          {{ error }}
        </div>

        <button class="btn btn-primary mr-2" @click="create">Create</button>
        <nuxt-link to="/admin/project">Cancel</nuxt-link>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      name: '',
      nameValidateError: '',
      error: ''
    }
  },
  methods: {
    async create() {
      const res = await this.$api.ProjectCreate(this.name)
      console.log('project create result: %o', res)
      if (!res.ok) {
        this.error = res.message
        return
      }

      await this.$bvModal.msgBoxOk('successfully created.')
      this.$router.push('/admin/project')
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
