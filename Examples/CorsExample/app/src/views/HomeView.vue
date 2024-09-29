<script setup>
import { toast } from "vue-sonner";
import axios from "axios";

const alert = async () => {
  try {
    const response = await axios.post("http://localhost:3000/hello");
    toast.success(response.data.message);
  } catch (error) {
    if (error.response) {
      toast.error(error.response.data.message);
    }
    toast.error(error.message);
  }
};
</script>

<template>
  <main>
    <div class="text-center">
      <p>If cors is configured properly, you should see a success toast</p>
      <button class="btn" @click="alert">Alert</button>
    </div>
  </main>
</template>
<style scoped>
main {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
}
.text-center {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100vw;
  gap: 1rem;
  font-size: 2rem;
}
.btn {
  padding: 0.5rem 1rem;
  font-size: large;
}
</style>
