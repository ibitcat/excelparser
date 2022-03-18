--   id                           int        任务id
--   name                         string     字符串
--   need                         int        定点数
--   need1                        json       定点数1

return {
  [1] = {
    id = 1,
    name = "获得\"传\"承",
    need = 1,
    need1 = {
      [1] = 1,
      [2] = 2,
      [3] = 3
    }
  },
  [2] = {
    id = 2,
    name = "获得传承",
    need = 1,
    need1 = {
      sites = {
        [1] = {
          name = "菜鸟教程",
          url = "www.runoob.com"
        },
        [2] = {
          name = "google",
          url = "www.google.com"
        },
        [3] = {
          url = "www.weibo.com",
          name = "微博"
        }
      }
    }
  }
}
