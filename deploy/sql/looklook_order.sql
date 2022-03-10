/*
 Navicat MySQL Data Transfer

 Source Server         : looklook
 Source Server Type    : MySQL
 Source Server Version : 80028
 Source Host           : 127.0.0.1:33069
 Source Schema         : looklook_order

 Target Server Type    : MySQL
 Target Server Version : 80028
 File Encoding         : 65001

 Date: 10/03/2022 17:15:38
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for homestay_order
-- ----------------------------
DROP TABLE IF EXISTS `homestay_order`;
CREATE TABLE `homestay_order` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `delete_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_state` tinyint NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0' COMMENT '版本号',
  `sn` char(25) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '订单号',
  `user_id` bigint NOT NULL DEFAULT '0' COMMENT '下单用户id',
  `homestay_id` bigint NOT NULL DEFAULT '0' COMMENT '民宿id',
  `title` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '标题',
  `sub_title` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '副标题',
  `cover` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '封面',
  `info` varchar(4069) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '介绍',
  `people_num` tinyint(1) NOT NULL DEFAULT '0' COMMENT '容纳人的数量',
  `row_type` tinyint(1) NOT NULL DEFAULT '0' COMMENT '售卖类型0：按房间出售 1:按人次出售',
  `need_food` tinyint(1) NOT NULL DEFAULT '0' COMMENT '0:不需要餐食 1:需要参数',
  `food_info` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '餐食标准',
  `food_price` bigint NOT NULL COMMENT '餐食价格(分)',
  `homestay_price` bigint NOT NULL COMMENT '民宿价格(分)',
  `market_homestay_price` bigint NOT NULL DEFAULT '0' COMMENT '民宿市场价格(分)',
  `homestay_business_id` bigint NOT NULL DEFAULT '0' COMMENT '店铺id',
  `homestay_user_id` bigint NOT NULL DEFAULT '0' COMMENT '店铺房东id',
  `live_start_date` date NOT NULL COMMENT '开始入住日期',
  `live_end_date` date NOT NULL COMMENT '结束入住日期',
  `live_people_num` tinyint NOT NULL DEFAULT '0' COMMENT '实际入住人数',
  `trade_state` tinyint(1) NOT NULL DEFAULT '0' COMMENT '-1: 已取消 0:待支付 1:未使用 2:已使用  3:已退款 4:已过期',
  `trade_code` char(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '确认码',
  `remark` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户下单备注',
  `order_total_price` bigint NOT NULL DEFAULT '0' COMMENT '订单总价格（餐食总价格+民宿总价格）(分)',
  `food_total_price` bigint NOT NULL DEFAULT '0' COMMENT '餐食总价格(分)',
  `homestay_total_price` bigint NOT NULL DEFAULT '0' COMMENT '民宿总价格(分)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_sn` (`sn`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='每一间民宿';

SET FOREIGN_KEY_CHECKS = 1;
