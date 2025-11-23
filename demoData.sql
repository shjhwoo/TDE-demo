-- --------------------------------------------------------
-- 호스트:                          localhost
-- 서버 버전:                        10.11.15-MariaDB-ubu2204 - mariadb.org binary distribution
-- 서버 OS:                        debian-linux-gnu
-- HeidiSQL 버전:                  11.3.0.6295
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


-- demoDB 데이터베이스 구조 내보내기
CREATE DATABASE IF NOT EXISTS `demoDB` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci */;
USE `demoDB`;

-- 테이블 demoDB.CreditCards 구조 내보내기
CREATE TABLE IF NOT EXISTS `CreditCards` (
  `Id` int(11) NOT NULL AUTO_INCREMENT,
  `OwnerName` varchar(50) DEFAULT NULL,
  `CardNumber` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`Id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- 테이블 데이터 demoDB.CreditCards:~2 rows (대략적) 내보내기
DELETE FROM `CreditCards`;
/*!40000 ALTER TABLE `CreditCards` DISABLE KEYS */;
INSERT INTO `CreditCards` (`Id`, `OwnerName`, `CardNumber`) VALUES
	(1, 'Cheolsu Kim', '1234-5678-9012-3456'),
	(2, 'Younghee Lee', '9876-5432-1098-7654');
/*!40000 ALTER TABLE `CreditCards` ENABLE KEYS */;

/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40111 SET SQL_NOTES=IFNULL(@OLD_SQL_NOTES, 1) */;
