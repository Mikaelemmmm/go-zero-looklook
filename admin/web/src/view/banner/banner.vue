<template>
    <div>
        banner <br />
    </div>
</template>

<script>
import {getBannerDetail} from "@/api/banner.js"
import {ElMessage} from "element-plus";
export default {
    name: "Test",
    created(){
     getBannerDetail({"id":1}).then((res)=>{
       ElMessage({
         showClose: true,
         message: "获取到了数据 ， id：" +res.data.id+"， title：" +res.data.Title,
         type: 'success'
       })
      })
    }
}
</script>
