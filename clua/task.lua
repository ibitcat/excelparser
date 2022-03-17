--   id                           int        任务id
--   name                         string     字符串
--   need                         int        定点数
--   need1                        json       定点数1
--   showProgress                 float      浮点数
--   dict                         dict<3>    字典
--     key1                       int        字段1
--     key2                       int        字段2
--     key3                       dict<2>    字典
--       key21                    int        字段1
--       key22                    int        字段2
--   list                         int[2][3]  基础数组
--     [1]                        int[2]     数组1
--       [1]                      int        字段1
--       [2]                      int        字段2
--     [2]                        int[2]     数组2
--       [1]                      int        字段1
--       [2]                      int        字段2
--     [3]                        int[2]     数组3
--       [1]                      int        字段1
--       [2]                      int        字段2
--   list1                        dict[2]    字典数组
--     [1]                        dict<2>    字段1
--       a                        int        字段2
--       b                        int        字段1
--     [2]                        dict<2>    字段1
--       a                        int        字段2
--       b                        int        字段1

return {
  [1] = {
    id = 1,
    name = "获得\"传\"承",
    need = 1,
    need1 = {
      [1] = 1,
      [2] = 2,
      [3] = 3
    },
    showProgress = 3.14,
    dict = {
      key1 = 3.14,
      key2 = 3.14,
      key3 = {
        key21 = 3.14,
        key22 = 3.14
      }
    },
    list = {
      [1] = {
        [1] = 1111,
        [2] = 2222
      },
      [2] = {
        [1] = 333,
        [2] = 444
      },
      [3] = {
        [1] = 333,
        [2] = 444
      }
    },
    list1 = {
      [1] = {
        a = 1,
        b = 2
      },
      [2] = {
        a = 1,
        b = 2
      }
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
          url = "www.google.com",
          name = "google"
        },
        [3] = {
          url = "www.weibo.com",
          name = "微博"
        }
      }
    },
    showProgress = 3.14,
    dict = {
      key1 = 3.14,
      key2 = 3.14,
      key3 = {
        key21 = 3.14,
        key22 = 3.14
      }
    },
    list = {
      [1] = {
        [1] = 1111,
        [2] = 2222
      },
      [2] = {
        [1] = 333,
        [2] = 444
      },
      [3] = {
        [1] = 333,
        [2] = 444
      }
    },
    list1 = {
      [1] = {
        a = 1,
        b = 2
      },
      [2] = {
        a = 1,
        b = 2
      }
    }
  }
}
